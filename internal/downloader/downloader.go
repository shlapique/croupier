package downloader

import (
	"context"
	"encoding/hex"
	// "sync"
	"crypto/md5"
	"net/http"
	"os"
	// "errors"
	"fmt"
	"io"
	"log/slog"
)

type cancelCmd int

const kill cancelCmd = 67

type Downloader struct {
	l *slog.Logger

	FilesChan   chan File
	workers     []*Worker
	MaxNumFiles int // at one time downloading
}

type Config struct {
	L            *slog.Logger
	DownloadPath string
	MaxNumFiles  int // at one time downloading
	WorkersNum   int
}

func New(ctx context.Context, config Config) *Downloader {
	l := config.L
	if l == nil {
		l = slog.New(slog.Default().Handler())
	}
	l = l.With("pkg", "downloader")

	workers := make([]*Worker, config.WorkersNum)
	filesChan := make(chan File, config.MaxNumFiles)

	// create workers
	for i := range config.WorkersNum {
		wLogger := l.With("worker", i)
		var w = createWorker(wLogger, i, filesChan, config.DownloadPath)
		go w.run(ctx)
		workers[i] = w
	}
	return &Downloader{
		FilesChan:   filesChan,
		workers:     workers,
		MaxNumFiles: config.MaxNumFiles,
	}
}

func (d *Downloader) CancelAll() {
	// consume all files from FilesChan
	// then cancel current worker files
	go func() {
		for {
			f, ok := <-d.FilesChan
			if !ok {
				d.l.Info("FilesChan closed, draining done")
				d.cancelWorkers()
				return
			}
			d.l.Info("removed file", "id", f.GetID())
		}
	}()
}

func (d *Downloader) cancelWorkers() {
	// d.mu.Lock()
	// defer d.mu.Unlock()
	busyFound := false
	for _, w := range d.workers {
		if w.Busy {
			busyFound = true
			d.l.Debug("Found Busy worker", "workerId", w.Id)
			w.Ctrl <- kill
		}
	}
	if !busyFound {
		d.l.Info("There's no workers with status Busy")
	}
}

// sum: md5sum
// if check sum provided: check sum of a file in filePath -> skip if already exists
func downloadFile(ctx context.Context, filePath string, url string, sum *string) error {
	if sum != nil {
		_, err := os.Stat(filePath)
		if !os.IsNotExist(err) {
			exists, err := checkSum(filePath, sum)
			if err != nil {
				fmt.Printf("Unable to check sum for file [%s], then downloading it ...\n", filePath)
			}
			if exists {
				// fmt.Printf("File [%s] exists! Skipping...\n", filePath)
				return nil
			}
		}
	}

	out, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Unable to create file [%s]\n", filePath)
		return err
	}
	defer out.Close()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// check sum post download
	_, err = checkSum(filePath, sum)
	if err != nil {
		return err
	}
	return nil
}

// comapre file by filePath with md5 string
func checkSum(filePath string, sum *string) (bool, error) {
	fileSum, err := md5Sum(filePath)
	if err != nil {
		return false, err
	}
	if fileSum == *sum {
		return true, nil
	}
	return false, nil
}

func md5Sum(filePath string) (string, error) {
	fi, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer fi.Close()

	h := md5.New()
	for {
		var data [256]byte
		n, err := fi.Read(data[:])
		if n == 0 && err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}
		h.Write(data[0:n])
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
