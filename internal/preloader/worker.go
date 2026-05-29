package preloader

import (
	"context"
	"fmt"
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
			v, err := w.fetch(job.offset)
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
