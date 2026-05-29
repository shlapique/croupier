package slider

import (
	// "context"
	// "sync"
	"fmt"
	"errors"
)

// two-wise 'overwriting' buffer without 'del' cmd
type ringBuffer[T any] struct {
	buffer   []*T

	ht       int // stands for "head-tail"

	count    int
	capacity int

	way       Way // current 'way' of moving
}

func newRingBuffer[T any](capacity int) *ringBuffer[T] {
	return &ringBuffer[T]{
		buffer:   make([]*T, capacity),
		ht:       0,
		count:    0,
		capacity: capacity,
		way:      Right,
	}
}

// moves tail one step forward newWay
// newWay == Left -> move h;t 1 step left
// newWay == Right -> move h;t 1 step right 
func (rb *ringBuffer[T]) changeWay(newWay Way) error {
	// krol jump to the opposite direction on 1 step
	if rb.way != newWay {
		switch newWay {
		case Left:
			rb.ht = (rb.ht  + rb.capacity - 1) % rb.capacity
			rb.way = newWay
			return nil
		case Right:
			rb.ht = (rb.ht + 1) % rb.capacity
			rb.way = newWay
			return nil
		default:
			return errors.New(fmt.Sprintf("unable to determine way: %v", newWay))
		} 
	} else {
		fmt.Println("current buffer way:", rb.way)
		fmt.Println("new way:", newWay, "-> nothing to do")
		return nil
	}
}

func (rb *ringBuffer[T]) add(el *T) {
	rb.buffer[rb.ht] = el 
	switch rb.way {
	case Left:
		rb.ht = (rb.ht + rb.capacity - 1) % rb.capacity
	case Right:
		rb.ht = (rb.ht + 1) % rb.capacity
	}
	if el == nil {
		fmt.Println("element:", el, "added!")
	} else {
		fmt.Println("element:", *el, "added!")
	}
	fmt.Println("New ht index:", rb.ht)
}
