package preloader 

import (
	// "context"
	// "sync"
	"fmt"
	"errors"

	"croupier/internal/slider"
)

// Fetcher receives the index and returns the data by index.
// It should be safe to call concurrently if the preloader uses goroutines.
type Fetcher[T any] func(index int) (T, error)

type Preloader[T any] struct {
	Sw     *slider.SlidingWindow[T]
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

	sw, err := slider.NewSlidingWindow[T](windowSize)
	if err != nil {
		fmt.Println("Unable to create NewSlidingWindow")
		return nil, err
	}

	loader := &Preloader[T]{
		Sw:        sw,
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
		fmt.Println("i:", i, "v:", v)
		data[i-l] = &v
	}

	loader.Sw.Init(data)

	return loader, nil
}

func (loader *Preloader[T]) LoadLeft(el *T) error {
	err := loader.Sw.SlideLeft(el)
	if err != nil {
		fmt.Println("unable to load left:", err)
		return err
	}
	return nil
}

func (loader *Preloader[T]) LoadRight(el *T) error {
	err := loader.Sw.SlideRight(el)
	if err != nil {
		fmt.Println("unable to load right:", err)
		return err
	}
	return nil
}
