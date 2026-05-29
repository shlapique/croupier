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
// l, r: left/right boundaries of i
type Fetcher[T any] func(index int, l int, r int) (T, error)

type Preloader[T any] struct {
	Sw     *slider.SlidingWindow[T]
	lag    int // a point (index) of SlidingWindow simmetry (or just a 'peephole')
	fetch  Fetcher[T]

	offset    int // real 'skew' offset (index) in real data that we work with
	minOffset int // 0
	maxOffset int // real maxOffset in data
}

func getLR(offset int, minOffset int, maxOffset int, lag int, windowSize int) (int, int) {
	return max(0, offset - lag), min(maxOffset, offset + (windowSize-1-lag))
}

func NewPreloader[T any](offset int, minOffset int, maxOffset int, windowSize int, lag int, fetcherFunc Fetcher[T]) (*Preloader[T], error) {
	if lag < 0 || lag > (windowSize-1) {
		fmt.Println("lag has to be insize [0, windowSize-1]")
		return nil, errors.New("incorrect Preloader lag")
	}
	if minOffset > maxOffset || maxOffset < minOffset {
		fmt.Println("minOffset and maxOffset have to be [minOffset, maxOffset]")
		return nil, errors.New(fmt.Sprintf("incorrect Preloader minOffset [%d] or maxOffset [%d]\n", minOffset, maxOffset))
	}
	if offset < minOffset || offset > maxOffset {
		fmt.Println("offset has to be insize [minOffset, maxOffset]")
		return nil, errors.New(fmt.Sprintf("incorrect Preloader offset [%d]\n", offset))
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
		minOffset: minOffset,
		maxOffset: maxOffset,
	}
	
	l, r := getLR(offset, minOffset, maxOffset, lag, windowSize)
	fmt.Println("L =", l, "R =", r)
	data := make([]*T, r-l+1)
	for i := l; i <= r; i++ {
		v, err := loader.fetch(i, loader.minOffset, loader.maxOffset)
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

func (loader *Preloader[T]) LoadLeft() error {
	if loader.offset == loader.minOffset {
		fmt.Printf("Unable to move more left then minOffset [%d], current offset [%d]", loader.minOffset, loader.offset)
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
		el, err := loader.fetch(idx, loader.minOffset, loader.maxOffset)
		if err != nil {
			fmt.Printf("Unable to fetch on index: %d\n", idx)
			return err
		}
		err = loader.Sw.SlideLeft(&el)
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
		el, err := loader.fetch(idx, loader.minOffset, loader.maxOffset)
		if err != nil {
			fmt.Printf("Unable to fetch on index: %d\n", idx)
			return err
		}
		err = loader.Sw.SlideRight(&el)
		if err != nil {
			fmt.Println("unable to load fight:", err)
			return err
		}
	}
	loader.offset += 1
	return nil
}
