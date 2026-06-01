package preloader

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Worker[T any] struct {
	id   int
	Offset int // current job.offset (if busy == true)

	jobs chan *Job[T]
	ctrl chan int  // offsets to kill

	minOffset int
	maxOffset int

	fetch 	    Fetcher[T]

	Busy bool
	// mu 
}

type fetchResult[T any] struct {
	v   T
	err error
}

func newWorker[T any](id int, jobChan chan *Job[T], fetchFunc Fetcher[T]) *Worker[T] {
	return &Worker[T]{
		id:   id,
		jobs: jobChan,
		ctrl: make(chan int, 10),
		fetch: fetchFunc,
	}
}

func (w *Worker[T]) run(ctx context.Context) {
	fmt.Printf("worker %d is running\n", w.id)

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("worker %d killed\n", w.id)
			return
		case job, ok := <-w.jobs:
			if !ok {
				fmt.Printf("worker %d: jobChan closed!\n", w.id)
				return
			}
			fmt.Printf("worker %d got new job: %d\n", w.id, job.offset)

			// get the item from data and assign where it belongs to
			// v, err := w.fetch(job.offset)
			v, err := w.timeoutFetch(ctx, job.offset)
			if err != nil {
				fmt.Println("Unable to fetch!:", err)
				break
			}
			fmt.Printf("worker %d got item: %s\n", w.id, v)
			fmt.Printf("worker %d, current v addr %v\n", w.id, &v)
			*job.el = v

		case offsetToKill := <-w.ctrl:
			fmt.Printf("worker %d got ctrl offset (to kill): %d\n", w.id, offsetToKill)
		}
	}
}

// set a timeout for fetch ALL opearations (e.g. 15s) -- sane
// fetch itself may have timeout too
// FIXME remove hardcoded 15 value
func (w *Worker[T]) timeoutFetch(ctx context.Context, i int) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// to return nil
	var zero T

	result := make(chan fetchResult[T], 1)

	go func() {
		v, err := w.fetch(i)
		result <- fetchResult[T]{v: v, err: err}
	}()

	select {
	case r := <- result:
		return r.v, r.err
	case <-ctx.Done():
		fmt.Println("Timeout [15s] for 'timeoutFetch' function exceeded!")
		return zero, errors.New("Timeout [15s] for 'timeoutFetch' function exceeded!")
	}
}
