package downloader

import (
	"context"
	// "errors"
	"fmt"
	// "time"
)

type Worker struct {
	Id int

	DownloadPath string

	HotFile File // current downloading file

	FilesChan chan File // job chan
	Ctrl  chan cancelCmd

	Busy bool
}

func createWorker(id int, filesChan chan File, DownloadPath string) *Worker {
	fmt.Printf("Created worker %s!\n", id)
	return &Worker{
		Id:           id,
		DownloadPath: DownloadPath,
		FilesChan:        filesChan,
		Ctrl:         make(chan cancelCmd, 100),
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
		case file, ok := <-w.FilesChan:
			if !ok {
				fmt.Printf("[Dworker %d] fileChan closed!\n", w.Id)
				return
			}
			fmt.Printf("[Dworker %d] got new file: %d\n", w.Id, file.GetID())

			w.Busy = true
			w.HotFile = file

			err := w.getFile(ctx, file)
			if err != nil {
				fmt.Printf("[Dworker %d] Unable to fetch!: %v\n", w.Id, err)
				w.Busy = false
				break
			}
			w.Busy = false
			fmt.Printf("[Dworker %d] Waiting for a new File\n", w.Id)
		}
	}
}

func (w *Worker) getFile(ctx context.Context, f File) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	result := make(chan error, 1)
	path := w.DownloadPath + f.GetName()
	link := f.GetHref()
	sum := f.GetMD5()

	go func() {
		fmt.Printf("[Dworker %d] downloading %d ...\n", w.Id, w.HotFile)
		err := downloadFile(ctx, path, link, sum)
		result <- err
	}()

	select {
	case r := <-result:
		if r != nil {
			return r
		}
		fmt.Printf("[Dworker %d] Downloaded file [%s]!\n", w.Id, f.GetName())
		return r

	case <-w.Ctrl:
		fmt.Printf("[Dworker %d] file download cancelled!\n", w.Id)
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}
