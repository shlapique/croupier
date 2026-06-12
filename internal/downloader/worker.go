package downloader

import (
	"context"
	// "errors"
	"fmt"
	// "time"

	// "github.com/google/uuid"
)

type Worker[T any] struct {
	Id int
	
	DownloadPath string

	HotFile *T // current downloading file 

	files chan *T // job chan
	Ctrl  chan int

	Busy bool
}

func createWorker[T any](id int, files chan *T, DownloadPath string) *Worker {
	fmt.Printf("Created worker %s!\n", id.String())
	return &Worker[T]{
		Id:    id,
		DownloadPath: DownloadPath,
		files: files,
		Ctrl:  make(chan int, 100),
	}
}

func (w *Worker) run(ctx context.Context) {
	fmt.Printf("[Dworker %d] running\n", w.Id)
	defer close(w.Ctrl)

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[Dworker %d] killed\n", w.Id)
			return
		case file, ok := <-w.files:
			if !ok {
				fmt.Printf("[Dworker %d] fileChan closed!\n", w.Id)
				return
			}
			fmt.Printf("[Dworker %d] got new file: %d\n", w.Id, *file.Id)

			w.Busy = true
			w.HotFile = file
			
			err := downloadFile(ctx, filePath, link, sum)
			// err := getFile(ctx, 

			if err != nil {
				fmt.Printf("[Dworker %d] Unable to fetch!: %v\n", w.Id, err)
				w.Busy = false
				break
			}
			if v != nil {
				fmt.Printf("[Dworker %d] got item: %s!\n", w.Id, *v)
				*job.el = *v
			} else {
				job.el = nil
			}
			w.Busy = false
			fmt.Printf("[Dworker %d] OK\n", w.Id)
	}
}

func (w *Worker) getFile(ctx context.Context, f *File) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	result := make(error, 1)
	path := w.DownloadPath + "/" + f.Name
	link := f.Href

	go func() {
		fmt.Printf("[Dworker %d] downloading %d ...\n", w.Id, w.HotFile)
		err := downloadFile(ctx, path, link, sum)
		result <- err
	}()

	select {
	case r := <-result
		return r.v, r.err

	case indexToKill := <-w.Ctrl:
		fmt.Printf("[Dworker %d] file download %d cancelled!\n", w.Id, indexToKill)
		return nil, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
