package preloader 

import (
	"context"
	// "sync"
	"fmt"
	"errors"

	"croupier/internal/slider"
)

// Fetcher receives the index and returns the data by index.
// It should be safe to call concurrently if the preloader uses goroutines.
// l, r: left/right boundaries of i
type Fetcher[T any] func(i int) (T, error)

type Preloader[T any] struct {
	Sw        *slider.SlidingWindow[T]

	jobChan   chan *Job[T]
	workers   []*Worker[T] // array of chan to communicate with workers

	offset    int // real 'skew' offset (index) in real data that we work with
	minOffset int // 0
	maxOffset int // real maxOffset in data
	lag       int // a point (index) of SlidingWindow simmetry (or just a 'peephole')
	fetch     Fetcher[T]
}

type Config[T any] struct {
	Offset    int // starting data index [MinOffset...MaxOffset]
	MinOffset int // 0 (usually)
	MaxOffset int // 

	Size int // i.e. windowSize

	// Lag === an index of simmetry (or just a 'peephole'):
	// when 0: we preload forward only ...
	// when 1: preload 1 back and others forward...
	Lag int 

	FetchFunc Fetcher[T]

	WorkersNum int
}

func getLR(offset int, minOffset int, maxOffset int, lag int, windowSize int) (int, int) {
	return max(0, offset - lag), min(maxOffset, offset + (windowSize-1-lag))
}

func New[T any](ctx context.Context, config Config[T]) (*Preloader[T], error) {
	if config.Lag < 0 || config.Lag > (config.Size-1) {
		fmt.Println("Lag has to be insize [0, Size-1]")
		return nil, errors.New("incorrect Preloader Lag")
	}
	if config.MinOffset > config.MaxOffset || config.MaxOffset < config.MinOffset {
		fmt.Println("MinOffset and MaxOffset have to be [MinOffset, MaxOffset]")
		return nil, errors.New(fmt.Sprintf("incorrect Preloader MinOffset [%d] or MaxOffset [%d]\n", config.MinOffset, config.MaxOffset))
	}
	if config.Offset < config.MinOffset || config.Offset > config.MaxOffset {
		fmt.Println("Offset has to be insize [MinOffset, MaxOffset]")
		return nil, errors.New(fmt.Sprintf("incorrect Preloader offset [%d]\n", config.Offset))
	}

	sw, err := slider.New[T](config.Size)
	if err != nil {
		fmt.Println("Unable to create NewSlidingWindow")
		return nil, err
	}

	workers := make([]*Worker[T], config.WorkersNum)
	// FIXME add jobs chan size param
	jobChan := make(chan *Job[T], 10)
	// create workers
	for i := range config.WorkersNum {
		var w = newWorker[T](i, jobChan, config.FetchFunc)
		go w.run(ctx)
		workers[i] = w
	}

	loader := &Preloader[T]{
		Sw:        sw,
		lag:       config.Lag,
		fetch:     config.FetchFunc,
		offset:    config.Offset,
		minOffset: config.MinOffset,
		maxOffset: config.MaxOffset,

		workers:   workers,
		jobChan:   jobChan,
	}

	fmt.Println("Initializing Preloader")
	l, r := getLR(loader.offset, loader.minOffset, loader.maxOffset, loader.lag, loader.Sw.Size)
	fmt.Println("L =", l, "R =", r)

	data := make([]*T, r-l+1)

	for i := l; i <= r; i++ {
		v := new(T)
		job := Job[T]{v, i}
		fmt.Printf("CREATING i: %d, v addr: %v\n", i, v)
		loader.jobChan <-&job
		data[i-l] = v
	}

	loader.Sw.Init(data)
	fmt.Println("OK")

	return loader, nil
}

func (loader *Preloader[T]) LoadLeft() error {
	if loader.offset == loader.minOffset {
		fmt.Printf("Unable to move more left then minOffset [%d], current offset [%d]\n", loader.minOffset, loader.offset)
		return errors.New(fmt.Sprintf("Unable to move more left then minOffset [%d], current offset [%d]\n", loader.minOffset, loader.offset))
	}

	// touching the edge of sw
	if loader.offset <= (loader.minOffset + loader.lag) {
		err := loader.Sw.SlideLeft(nil)
		if err != nil {
			fmt.Println("unable to load left:", err)
			return err
		}
	} else {
		idx := loader.offset - loader.lag - 1
		v := new(T)
		job := Job[T]{v, idx}
		loader.jobChan <-&job

		err := loader.Sw.SlideLeft(v)
		if err != nil {
			fmt.Println("unable to load left:", err)
			return err
		}
	}
	loader.offset -= 1
	return nil
}

func (loader *Preloader[T]) LoadRight() error {
	if loader.offset == loader.maxOffset {
		fmt.Printf("Unable to move more right then maxOffset [%d], current offset [%d]\n", loader.maxOffset, loader.offset)
		return errors.New(fmt.Sprintf("Unable to move right then maxOffset [%d], current offset [%d]\n", loader.maxOffset, loader.offset))
	}

	// touching the edge of sw
	if loader.offset >= (loader.maxOffset - (loader.Sw.Size-1-loader.lag)) {
		err := loader.Sw.SlideRight(nil)
		if err != nil {
			fmt.Println("unable to load right:", err)
			return err
		}
	} else {
		idx := loader.offset + loader.Sw.Size - loader.lag 
		v := new(T)
		job := Job[T]{v, idx}
		loader.jobChan <-&job

		err := loader.Sw.SlideRight(v)
		if err != nil {
			fmt.Println("unable to load fight:", err)
			return err
		}
	}
	loader.offset += 1
	return nil
}
