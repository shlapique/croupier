package preloader

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Worker[T any] struct {
	Id   int
	Offset int // current job.offset (if busy == true)

	jobs chan *Job[T]
	Ctrl chan int  // offsets to kill

	minOffset int
	maxOffset int

	fetch 	    Fetcher[T]

	Busy bool
	// mu 
}

type fetchResult[T any] struct {
	v   *T
	err error
}

func newWorker[T any](id int, jobChan chan *Job[T], fetchFunc Fetcher[T]) *Worker[T] {
	return &Worker[T]{
		Id:   id,
		jobs: jobChan,
		Ctrl: make(chan int, 100),
		fetch: fetchFunc,
	}
}

func (w *Worker[T]) run(ctx context.Context) {
	fmt.Printf("[worker %d] running\n", w.Id)
	defer close(w.Ctrl)

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[worker %d] killed\n", w.Id)
			return
		case job, ok := <-w.jobs:
			if !ok {
				fmt.Printf("[worker %d] jobChan closed!\n", w.Id)
				return
			}
			fmt.Printf("[worker %d] got new job: %d\n", w.Id, job.offset)

			// get the item from data and assign where it belongs to
			w.Busy = true
			w.Offset = job.offset
			v, err := w.timeoutFetch(ctx, job.offset)
			if err != nil {
				fmt.Printf("[worker %d] Unable to fetch!: %v\n", w.Id, err)
				w.Busy = false
				break
			}
			if v != nil {
				fmt.Printf("[worker %d] got item: %s!\n", w.Id, *v)
				*job.el = *v
			} else {
				job.el = nil
			}
			w.Busy = false
			fmt.Printf("[worker %d] OK\n", w.Id)

		case offsetToKill := <-w.Ctrl:
			fmt.Printf("[worker %d] got ctrl offset (to kill): %d\n", w.Id, offsetToKill)
		}
	}
}

// set a timeout for fetch ALL opearations (e.g. 15s) -- sane
// fetch itself may have timeout too
// FIXME remove hardcoded 15 value
func (w *Worker[T]) timeoutFetch(ctx context.Context, i int) (*T, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	result := make(chan fetchResult[T], 1)

	go func() {
		fmt.Printf("[worker %d] Fetching %d ...\n", w.Id, i)
		v, err := w.fetch(i)
		result <- fetchResult[T]{v: &v, err: err}
	}()

	select {
	case r := <- result:
		return r.v, r.err

	case indexToKill := <-w.Ctrl:
		fmt.Printf("[worker %d] job %d cancelled!\n", w.Id, indexToKill)
		return nil, nil

	case <-ctx.Done():
		fmt.Printf("[worker %d] Timeout [15s] for 'timeoutFetch' function exceeded!\n", w.Id)
		return nil, errors.New(fmt.Sprintf("[worker %d] Timeout [15s] for 'timeoutFetch' function exceeded!\n", w.Id))
	}
}
