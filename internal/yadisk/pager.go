package yadisk

import (
	// "context"
	// "sync"
	"fmt"
	"errors"
)

type Way int

const (
	Right Way = iota
	Left
)

// two-wise 'overwriting' buffer without 'del' cmd
type RingBuffer[T any] struct {
	buffer   []*T

	ht       int // stands for "head-tail"

	count    int
	capacity int

	way       Way // current 'way' of moving
}

type SlidingWindow[T any] struct {
	rb   *RingBuffer[T]
	size int 
}

// Fetcher receives the index and returns the data by index.
// It should be safe to call concurrently if the preloader uses goroutines.
type Fetcher[T any] func(index int) (T, error)

type Preloader[T any] struct {
	sw     *SlidingWindow[T]
	lag    int // a point (index) of SlidingWindow simmetry (or just a 'peephole')
	fetch  Fetcher[T]

	offset    int // real 'skew' offset (index) in real data that we work with
	maxOffset int // real maxOffset in data
}

// minOffset = 0
func getLR(offset int, maxOffset int, lag int, windowSize int) (int, int) {
	return max(0, offset - lag), min(maxOffset, offset + (windowSize-1-lag))
}

func NewPreloader[T any](offset int, maxOffset int, windowSize int, lag int, fetcherFunc Fetcher[T]) (*Preloader[T], error) {
	if lag < 0 || lag > (windowSize-1) {
		fmt.Println("lag has to be insize [0, windowSize-1]")
		return nil, errors.New("incorrect Preloader lag")
	}
	if maxOffset < 0 {
		fmt.Println("maxOffset has to be insize [0, ...]")
		return nil, errors.New("incorrect Preloader maxOffset")
	}
	if offset < 0 || offset > maxOffset {
		fmt.Println("offset has to be insize [0, maxOffset]")
		return nil, errors.New("incorrect Preloader offset")
	}

	sw, err := NewSlidingWindow[T](windowSize)
	if err != nil {
		fmt.Println("Unable to create NewSlidingWindow")
		return nil, err
	}

	loader := &Preloader[T]{
		sw:        sw,
		lag:       lag,
		fetch:     fetcherFunc,
		offset:    offset,
		maxOffset: maxOffset,
	}
	
	l, r := getLR(offset, maxOffset, lag, windowSize)
	fmt.Println("L =", l, "R =", r)
	data := make([]*T, r-l+1)
	for i := l; i <= r; i++ {
		v, err := loader.fetch(i)
		if err != nil {
			fmt.Println("Unable to fetch!:", err)
			break
		}
		// if v == nil {
		// 	fmt.Println("FOUND NIL (i =", i, ")")
		// 	break
		// }
		fmt.Println("i:", i, "v:", v)
		data[i-l] = &v
	}
	// FIXME preload task
	// preload 0+offset'th element +
	//   + (lag + 1 - windowSize) elements to the right
	loader.sw.init(data)
	return loader, nil
}

func NewSlidingWindow[T any](size int) (*SlidingWindow[T], error) {
	if size <= 1 {
		fmt.Println("Are you stupid.. sliding window with size", size, "?! hell na")
		return nil, errors.New(fmt.Sprintf("Are you stupid.. sliding window with size %d?! hell na", size))
	}

	sw := &SlidingWindow[T]{
		rb:   NewRingBuffer[T](size),
		size: size,
	}
	return sw, nil
}

// some els in data may be <nil>
func (sw *SlidingWindow[T]) init(data []*T) {
	fmt.Println("Intiating SlidingWindow with data")
	for _, v := range data {
		sw.rb.add(v)
	}
}

func NewRingBuffer[T any](capacity int) *RingBuffer[T] {
	return &RingBuffer[T]{
		buffer:   make([]*T, capacity),
		ht:       0,
		count:    0,
		capacity: capacity,
		way:      Right,
	}
}

func (sw *SlidingWindow[T]) slideLeft(el *T) error {
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

func (sw *SlidingWindow[T]) slideRight(el *T) error {
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

func (loader *Preloader[T]) LoadLeft(el *T) error {
	err := loader.sw.slideLeft(el)
	if err != nil {
		fmt.Println("unable to load left:", err)
		return err
	}
	return nil
}

func (loader *Preloader[T]) LoadRight(el *T) error {
	err := loader.sw.slideRight(el)
	if err != nil {
		fmt.Println("unable to load right:", err)
		return err
	}
	return nil
}

func (loader *Preloader[T]) ShowWindow() {
	for i, v := range loader.sw.rb.buffer {
		if v == nil {
			fmt.Println("i:", i, "v:", nil)
		} else {
			fmt.Println("i:", i, "v:", *v)
		}
	}
}

// moves tail one step forward newWay
// newWay == Left -> move h;t 1 step left
// newWay == Right -> move h;t 1 step right 
func (rb *RingBuffer[T]) changeWay(newWay Way) error {
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

func (rb *RingBuffer[T]) add(el *T) {
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
