package preloader

import (
	"context"
	// "sync"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"croupier/internal/slider"
)

// Fetcher receives the index and returns the data by index.
// It should be safe to call concurrently if the preloader uses goroutines.
// l, r: left/right boundaries of i
type Fetcher[T any] func(i int) (T, error)

type Preloader[T any] struct {
	l *slog.Logger

	Sw     *slider.SlidingWindow[T]
	Offset int  // real 'skew' offset (index) in real data that we work with
	Lag    int  // a point (index) of SlidingWindow simmetry (or just a 'peephole')
	Active bool // false -> call preloader.Init(); else -> do nothing

	jobChan chan *Job[T]
	workers []*Worker[T] // array of chan to communicate with workers

	minOffset int // 0
	maxOffset int // real maxOffset in data
	fetch     Fetcher[T]
}

type Config[T any] struct {
	L *slog.Logger

	Offset    int // starting data index [MinOffset...MaxOffset]
	MinOffset int // 0 (usually)
	MaxOffset int //

	Size int // i.e. windowSize
	// Lag === an index of simmetry (or just a 'peephole'):
	// when 0: we preload forward only ...
	// when 1: preload 1 back and others forward...
	Lag int

	FetchFunc Fetcher[T]
	Timeout   time.Duration

	WorkersNum int
}

func getLR(offset int, minOffset int, maxOffset int, lag int, windowSize int) (int, int) {
	return max(minOffset, offset-lag), min(maxOffset, offset+(windowSize-1-lag))
}

func New[T any](ctx context.Context, config Config[T]) (*Preloader[T], error) {
	l := config.L
	if l == nil {
		l = slog.New(slog.Default().Handler())
	}
	l = l.With("pkg", "loader")

	if config.Lag < 0 || config.Lag > (config.Size-1) {
		l.Error("Lag has to be insize [0, Size-1]")
		return nil, errors.New("incorrect Preloader Lag")
	}
	if config.MinOffset > config.MaxOffset || config.MaxOffset < config.MinOffset {
		l.Error("MinOffset and MaxOffset have to be [MinOffset, MaxOffset]")
		return nil, errors.New(fmt.Sprintf("incorrect Preloader MinOffset [%d] or MaxOffset [%d]\n", config.MinOffset, config.MaxOffset))
	}
	if config.Offset < config.MinOffset || config.Offset > config.MaxOffset {
		l.Error("Offset has to be insize [MinOffset, MaxOffset]")
		return nil, errors.New(fmt.Sprintf("incorrect Preloader offset [%d]\n", config.Offset))
	}

	sw, err := slider.New[T](config.Size)
	if err != nil {
		l.Error("Unable to create NewSlidingWindow", "err", err)
		return nil, err
	}

	workers := make([]*Worker[T], config.WorkersNum)
	// FIXME cfg
	jobChan := make(chan *Job[T], 500)
	// create workers
	for i := range config.WorkersNum {
		wLogger := l.With("worker", i)
		var w = newWorker[T](wLogger, i, jobChan, config.FetchFunc, config.Timeout)
		go w.run(ctx)
		workers[i] = w
	}

	loader := &Preloader[T]{
		l:         l,
		Sw:        sw,
		Lag:       config.Lag,
		fetch:     config.FetchFunc,
		Offset:    config.Offset,
		minOffset: config.MinOffset,
		maxOffset: config.MaxOffset,
		workers:   workers,
		jobChan:   jobChan,
	}

	return loader, nil
}

// call this function before any slider movements
func (loader *Preloader[T]) Init() {
	if loader.Active {
		loader.l.Info("Preloader already active. Skipping.")
		return
	}

	loader.l.Info("Initializing Preloader")
	l, r := getLR(loader.Offset, loader.minOffset, loader.maxOffset, loader.Lag, loader.Sw.Size)
	loader.l.Debug("L", l, "R", r)

	data := make([]*T, r-l+1)

	for i := l; i <= r; i++ {
		v := new(T)
		job := Job[T]{v, i}
		loader.l.Debug("creating i", "i", i, "&v", v)
		loader.jobChan <- &job
		data[i-l] = v
	}

	loader.Sw.Init(data)
}

func (loader *Preloader[T]) killJob(offsetIndex int) {
	for _, w := range loader.workers {
		if w.Busy && w.Offset == offsetIndex {
			loader.l.Debug("Found Busy worker with proper offset", "worker", w.Id, "offset", w.Offset)
			loader.l.Debug("Sending offsetIndex to kill", "offsetIndex", offsetIndex)
			w.Ctrl <- offsetIndex
			return
		}
	}
	loader.l.Info("There's no workers with proper needed offset and status Busy", "offset", offsetIndex)
	return
}

func (loader *Preloader[T]) LoadLeft() error {
	if loader.Offset == loader.minOffset {
		loader.l.Error("Unable to move more left then minOffset", "minOffset", loader.minOffset, "current offset", loader.Offset)
		return errors.New(fmt.Sprintf("Unable to move more left then minOffset [%d], current offset [%d]\n", loader.minOffset, loader.Offset))
	}

	// touching the edge of sw
	if loader.Offset <= (loader.minOffset + loader.Lag) {
		err := loader.Sw.SlideLeft(nil)
		if err != nil {
			loader.l.Error("Unable to load left", "err", err)
			return err
		}
	} else {
		// calc new index
		l, r := getLR(loader.Offset, loader.minOffset, loader.maxOffset, loader.Lag, loader.Sw.Size)
		idx := l - 1

		// cancel oldest (right) job
		loader.killJob(r)

		// assign new job
		v := new(T)
		job := Job[T]{v, idx}
		loader.jobChan <- &job

		err := loader.Sw.SlideLeft(v)
		if err != nil {
			loader.l.Error("Unable to load left", "err", err)
			return err
		}
	}
	loader.Offset -= 1
	return nil
}

func (loader *Preloader[T]) LoadRight() error {
	if loader.Offset == loader.maxOffset {
		loader.l.Error("Unable to move more right then maxOffset", "maxOffset", loader.maxOffset, "current offset", loader.Offset)
		return errors.New(fmt.Sprintf("Unable to move right then maxOffset [%d], current offset [%d]\n", loader.maxOffset, loader.Offset))
	}

	// touching the edge of sw
	if loader.Offset >= (loader.maxOffset - (loader.Sw.Size - 1 - loader.Lag)) {
		err := loader.Sw.SlideRight(nil)
		if err != nil {
			loader.l.Error("Unable to load right", "err", err)
			return err
		}
	} else {
		// calc new index
		l, r := getLR(loader.Offset, loader.minOffset, loader.maxOffset, loader.Lag, loader.Sw.Size)
		idx := r + 1

		// cancel oldest (left) job
		loader.killJob(l)

		// assign new job
		v := new(T)
		job := Job[T]{v, idx}
		loader.jobChan <- &job

		err := loader.Sw.SlideRight(v)
		if err != nil {
			loader.l.Error("Unable to load right", "err", err)
			return err
		}
	}
	loader.Offset += 1
	return nil
}

func (loader *Preloader[T]) ShowWindow() {
	for i := 0; i < loader.Sw.Size; i++ {
		v, err := loader.Sw.GetCell(i)
		if err != nil {
			loader.l.Error("Unable to get cell from Window", "cell", i)
			return
		}
		p := ""
		if i == loader.Lag {
			p = "<- LAG"
		}
		if v == nil {
			fmt.Printf("i: %d, v: %v %s\n", i, nil, p)
		} else {
			fmt.Printf("i: %d, v: %v %s\n", i, *v, p)
		}
	}
}
