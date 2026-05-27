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
type RingBuffer struct {
	buffer   []*interface{}

	ht       int // stands for "head-tail"

	count    int
	capacity int

	way        Way // current 'way' of moving
}

type SlidingWindow struct {
	rb     *RingBuffer

	offset     int // i.e. current page
	maxOffset  int // i.e. max num of pages
	windowSize int 
}

func (sw *SlidingWindow) slideLeft(element *interface{}) error {
	way := Left
	if sw.offset > 0 {
		if sw.rb.way != way {
			err := sw.rb.changeWay(way)
			if err != nil {
				fmt.Println("failed to change way:", err)
				return err
			}
		}
		sw.rb.add(element)
		return nil
	} else {
		return errors.New(fmt.Sprintf("Unable to move left. Offset = %d", sw.offset))
	}
}

func (sw *SlidingWindow) slideRight(element *interface{}) error {
	way := Right
	if sw.offset < sw.maxOffset {
		if sw.rb.way != way {
			err := sw.rb.changeWay(way)
			if err != nil {
				fmt.Println("failed to change way:", err)
				return err
			}
		}
		sw.rb.add(element)
		return nil
	} else {
		return errors.New(fmt.Sprintf("Unable to move right. Offset = %d", sw.offset))
	}
}

// Preloader and SlidingWindow have to know 
// max number of pages
type Preloader struct {
	sw        *SlidingWindow

	lag       int // 'peephole' index in window
	maxNumPages  int // for sw === maxOffset
}

func NewPreloader(maxNumPages int, windowSize int, lag int) (*Preloader, error) {
	if lag < 0 || lag > (windowSize-1) {
		fmt.Println("lag has to be insize [0, windowSize-1]")
		return nil, errors.New("incorrect Preloader lag")
	}
	sw, err := NewSlidingWindow(windowSize, maxNumPages)
	if err != nil {
		fmt.Println("Unable to create NewSlidingWindow")
		return nil, err
	}
	loader := &Preloader{
		sw:          sw,
		lag:         lag,
		maxNumPages: maxNumPages,
	}
	return loader, nil
}

func (loader *Preloader) LoadLeft(element *interface{}) error {
	if loader.sw.offset <= 0 {
		return fmt.Errorf("unable to move left: offset = %d", loader.sw.offset)
	}

	var err error

	if loader.sw.offset <= loader.lag {
		if element != nil {
			fmt.Println("trying to preload element (page) with implicit index < 0")
		} else {
			fmt.Println("element == nil -> cleaning (nilling) element from window")
		}
		err = loader.sw.slideLeft(nil)
	} else {
		err = loader.sw.slideLeft(element)
	}

	if err != nil {
		fmt.Println("unable to slide left:", err)
	}
	return err
}

func (loader *Preloader) LoadRight(element *interface{}) error {
	if loader.sw.offset >= loader.maxNumPages-1 {
		return fmt.Errorf("unable to move right: offset = %d", loader.sw.offset)
	}

	var err error

	if loader.sw.offset >= (loader.maxNumPages-1-loader.lag) {
		if element != nil {
			fmt.Println("trying to preload element (page) with implicit index > maxNumPages-1")
		} else {
			fmt.Println("element == nil -> cleaning (nilling) element from window")
		}
		err = loader.sw.slideRight(nil)
	} else {
		err = loader.sw.slideRight(element)
	}

	if err != nil {
		fmt.Println("unable to slide right:", err)
	}
	return err
}

func NewSlidingWindow(size int, maxOffset int) (*SlidingWindow, error) {
	if size <= 1 {
		fmt.Println("Are you stupid.. sliding window with size", size, "?! hell na")
		return nil, errors.New(fmt.Sprintf("Are you stupid.. sliding window with size %d?! hell na", size))
	}

	if maxOffset <= 1 {
		fmt.Println("Are you stupid.. maxOffset (maxNumPages) with size", maxOffset, "?! hell na")
		return nil, errors.New(fmt.Sprintf("Are you stupid.. maxOffset (maxNumPages) with size %d?! hell na", maxOffset))
	}
	sw := &SlidingWindow{
		rb:         NewRingBuffer(size),
		maxOffset:  maxOffset,
		windowSize: size,
	}
	return sw, nil
}

func NewRingBuffer(capacity int) *RingBuffer {
	buffer := make([]*interface{}, capacity)
	return &RingBuffer{
		buffer:   buffer,
		ht:       0,
		count:    0,
		capacity: capacity,
		way:      Right,
	}
}

// moves tail one step forward newWay
// newWay == Left -> move h;t 1 step left
// newWay == Right -> move h;t 1 step right 
func (rb *RingBuffer) changeWay(newWay Way) error {
	// krol jump to the opposite direction on 1 step
	if rb.way != newWay {
		switch newWay {
		case Left:
			rb.ht = (rb.ht  + rb.capacity - 1) % rb.capacity
			return nil
		case Right:
			rb.ht = (rb.ht + 1) % rb.capacity
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

func (rb *RingBuffer) add(element *interface{}) {
	rb.buffer[rb.ht] = element
	switch rb.way {
	case Left:
		rb.ht = (rb.ht  + rb.capacity - 1) % rb.capacity
	case Right:
		rb.ht = (rb.ht + 1) % rb.capacity
	}
	fmt.Println("element:", element, "added!")
	fmt.Println("New ht index:", rb.ht)
}

// func (rb *RingBuffer) get()

// type Pager struct {
// 	client *Client
// 	path   string
// 	limit  int
// }
