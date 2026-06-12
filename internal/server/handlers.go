package server

import (
	"net/http"
	// "context"
	// "sync"
	// "errors"
	"encoding/json"
	"fmt"

	"croupier/internal/preloader"
	// "croupier/internal/downloader"
	// "croupier/internal/yadisk"
)

type PreloaderHandlers[T any] struct {
	loader *preloader.Preloader[T]
}

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

func downloadSelectedHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("downloadSelectedHandler!\n")
}

func selectItemHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("selectItemHandler!\n")
}
