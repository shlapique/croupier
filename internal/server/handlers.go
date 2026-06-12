package server

import (
	"net/http"
	// "context"
	// "sync"
	// "errors"
	"fmt"
	"encoding/json"
	
	"croupier/internal/preloader"
	// "croupier/internal/downloader"
	// "croupier/internal/yadisk"
)

type stateHandler[T any] struct {
	loader *preloader.Preloader[T]
}

func (h stateHandler[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from a handler with a state!")
	v, err := h.loader.Sw.GetCell(h.loader.Lag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(v); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
	// fmt.Fprintf(w, "%+v\n", v)
	// firstFile := v.Files[0]
	// w.Write(v)
}

func getStateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("getStateHandler!\n")
}

func nextHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("nextHandler!\n")
}

func prevHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("prevHandler!\n")
}

func downloadSelectedHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("downloadSelectedHandler!\n")
}

func selectItemHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("selectItemHandler!\n")
}
