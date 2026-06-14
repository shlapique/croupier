package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	// "errors"
	"fmt"

	"croupier/internal/downloader"
	"croupier/internal/preloader"
	"croupier/internal/yadisk"
)

type PreloaderHandlers[T any] struct {
	loader *preloader.Preloader[T]
}

type DownloaderHandlers struct {
	mu         sync.Mutex
	cancel     context.CancelFunc
	downloader *downloader.Downloader
	client     *yadisk.Client
}

// maybe add current lag (offset) to response
func (h *PreloaderHandlers[T]) Current(w http.ResponseWriter, r *http.Request) {
	v, err := h.loader.Sw.GetCell(h.loader.Lag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *PreloaderHandlers[T]) Next(w http.ResponseWriter, r *http.Request) {
	err := h.loader.LoadRight()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *PreloaderHandlers[T]) Prev(w http.ResponseWriter, r *http.Request) {
	err := h.loader.LoadLeft()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *DownloaderHandlers) Download(w http.ResponseWriter, r *http.Request) {
	var files []yadisk.File
	if err := json.NewDecoder(r.Body).Decode(&files); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	defer r.Body.Close()

	if len(files) > (h.downloader.MaxNumFiles - len(h.downloader.FilesChan)) {
		http.Error(w, fmt.Sprintf("Supports up to %d in queue: available slots = %d\n",
			h.downloader.MaxNumFiles,
			h.downloader.MaxNumFiles-len(h.downloader.FilesChan)),
			http.StatusRequestEntityTooLarge)
		return
	}

	fmt.Printf("Got files to download:\n")
	fmt.Printf("%v", files)

	h.mu.Lock()
	if h.cancel != nil {
		h.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel
	h.mu.Unlock()

	// go block where we 'll obtain links to all files
	go func() {
		for _, v := range files {
			file := v
			resp, err := h.client.GetDownloadLink(ctx, file)

			if err != nil {
				fmt.Printf("Unable to get download link for a file: %s -> Skipping\n", file.Id)
				continue
			}
			file.Href = resp.Href

			select {
			case <-ctx.Done():
				return
			case h.downloader.FilesChan <- &file:
				// TODO remove
				fmt.Printf("added file [%s] to file chan\n", file.Id)
			}
		}
	}()

	w.WriteHeader(http.StatusOK)
}

func (h *DownloaderHandlers) CancelDownloads(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	if h.cancel != nil {
		h.cancel()
		h.cancel = nil
	}
	h.mu.Unlock()

	h.downloader.CancelAll()
	fmt.Printf("CancelDownloads fired!\n")
	w.WriteHeader(http.StatusOK)
}
