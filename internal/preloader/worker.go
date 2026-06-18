package preloader

import (
	"context"
	// "errors"
	// "fmt"
	"log/slog"
	"time"
)

type Worker[T any] struct {
	l *slog.Logger

	Id     int
	Offset int // current job.offset (if busy == true)

	jobs chan *Job[T]
	Ctrl chan int // offsets to kill

	minOffset int
	maxOffset int

	fetch   Fetcher[T]
	timeout time.Duration

	Busy bool
	// mu
}

type fetchResult[T any] struct {
	v   *T
	err error
}

func newWorker[T any](l *slog.Logger, id int, jobChan chan *Job[T], fetchFunc Fetcher[T], timeout time.Duration) *Worker[T] {
	return &Worker[T]{
		l:       l,
		Id:      id,
		jobs:    jobChan,
		Ctrl:    make(chan int, 100),
		fetch:   fetchFunc,
		timeout: timeout,
	}
}

func (w *Worker[T]) run(ctx context.Context) {
	w.l.Info("running")
	defer close(w.Ctrl)

	for {
		select {
		case <-ctx.Done():
			w.l.Info("killed")
			return
		case job, ok := <-w.jobs:
			if !ok {
				w.l.Info("jobChan closed!")
				return
			}
			w.l.Info("got new job", "job.offset", job.offset)

			// get the item from data and assign where it belongs to
			w.Busy = true
			w.Offset = job.offset
			v, err := w.timeoutFetch(ctx, job.offset)
			if err != nil {
				w.l.Error("Unable to fetch!", "err", err)
				w.Busy = false
				continue
			}
			if v != nil {
				w.l.Debug("got item", "value", *v)
				*job.el = *v
			} else {
				job.el = nil
			}
			w.Busy = false
			w.l.Info("Waiting for a new job")
		}
	}
}

func (w *Worker[T]) timeoutFetch(ctx context.Context, i int) (*T, error) {
	ctx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	result := make(chan fetchResult[T], 1)

	go func() {
		w.l.Debug("Fetching", "i", i)
		v, err := w.fetch(i)
		result <- fetchResult[T]{v: &v, err: err}
	}()

	select {
	case r := <-result:
		return r.v, r.err

	case indexToKill := <-w.Ctrl:
		w.l.Info("job cancelled!", "index", indexToKill)
		return nil, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
