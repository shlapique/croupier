package downloader

import (
	"context"
	// "errors"
	// "fmt"
	// "time"
	"log/slog"
)

type Worker struct {
	l *slog.Logger

	Id           int
	DownloadPath string

	HotFile File // current downloading file

	FilesChan chan File // job chan
	Ctrl      chan cancelCmd

	Busy bool
}

func createWorker(l *slog.Logger, id int, filesChan chan File, DownloadPath string) *Worker {
	return &Worker{
		l:            l,
		Id:           id,
		DownloadPath: DownloadPath,
		FilesChan:    filesChan,
		// FIXME cfg
		Ctrl: make(chan cancelCmd, 100),
	}
}

func (w *Worker) run(ctx context.Context) {
	w.l.Info("running")
	defer close(w.Ctrl)

	for {
		select {
		case <-ctx.Done():
			w.l.Info("killed")
			return
		case file, ok := <-w.FilesChan:
			if !ok {
				w.l.Info("fileChan closed!")
				return
			}
			w.l.Info("got new file", "file_Id", file.GetID())

			w.Busy = true
			w.HotFile = file

			err := w.getFile(ctx, file)
			if err != nil {
				w.l.Error("Unable to fetch!", "err", err)
				w.Busy = false
				break
			}
			w.Busy = false
			w.l.Info("Waiting for a new file")
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
		w.l.Debug("downloading", "file", w.HotFile)
		err := downloadFile(ctx, path, link, sum)
		result <- err
	}()

	select {
	case r := <-result:
		if r != nil {
			return r
		}
		w.l.Info("Downloaded file", "file", f.GetName())
		return r

	case <-w.Ctrl:
		w.l.Info("file download cancelled!")
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}
