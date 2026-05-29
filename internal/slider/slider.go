package slider

import (
	// "context"
	// "sync"
	"fmt"
	"errors"
)

type SlidingWindow[T any] struct {
	rb   *ringBuffer[T]
	Size int 
}

func New[T any](size int) (*SlidingWindow[T], error) {
	if size <= 1 {
		fmt.Println("Are you stupid.. sliding window with size", size, "?! hell na")
		return nil, errors.New(fmt.Sprintf("Are you stupid.. sliding window with size %d?! hell na", size))
	}

	sw := &SlidingWindow[T]{
		rb:   newRingBuffer[T](size),
		Size: size,
	}
	return sw, nil
}

// some els in data may be <nil>
func (sw *SlidingWindow[T]) Init(data []*T) {
	fmt.Println("Intiating SlidingWindow with data")
	for _, v := range data {
		sw.rb.add(v)
	}
}

func (sw *SlidingWindow[T]) SlideLeft(el *T) error {
	way := Left
	if sw.rb.way != way {
		err := sw.rb.changeWay(way)
		if err != nil {
			fmt.Println("failed to change way:", err)
			return err
		}
	}
	sw.rb.add(el)
	return nil
}

func (sw *SlidingWindow[T]) SlideRight(el *T) error {
	way := Right
	if sw.rb.way != way {
		err := sw.rb.changeWay(way)

		if err != nil {
			fmt.Println("failed to change way:", err)
			return err
		}
	}
	sw.rb.add(el)
	return nil
}

func (sw *SlidingWindow[T]) Show() {
	for i, v := range sw.rb.buffer {
		p := ""
		if i == sw.rb.ht {
			p = "<- ht"
		}
		if v == nil {
			fmt.Printf("i: %d, v: %v %s\n", i, nil, p)
		} else {
			fmt.Printf("i: %d, v: %v %s\n", i, *v, p)
		}
	}
}
