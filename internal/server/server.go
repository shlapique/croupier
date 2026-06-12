package server

// server works with some 'struct Response'
// and with some Preloader[T]

import (
	"context"
	"net/http"
	// "sync"
	// "errors"
	"fmt"
	"log"
	"io/fs"
	"embed"

	"croupier/internal/preloader"
)

type Server[T any] struct {
	srv    *http.Server
	Loader *preloader.Preloader[T]
	Port   string
}

func New[T any](loader *preloader.Preloader[T], port string, assets embed.FS) *Server[T] {
	mux := http.NewServeMux()

	sub, err := fs.Sub(assets, "static")
	if err != nil {
		panic(err)
	}

	mux.Handle("/", http.FileServer(http.FS(sub)))

	h := &PreloaderHandlers[T]{
		loader: loader,
	}

	mux.HandleFunc("GET /state", h.Current)
	mux.HandleFunc("POST /next", h.Next)
	mux.HandleFunc("POST /prev", h.Prev)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	return &Server[T]{
		srv:    srv,
		Loader: loader,
		Port:   port,
	}
}

func (server *Server[T]) Run(ctx context.Context) {
	fmt.Printf("server listening on %s\n", server.Port)

	go func() {
		<-ctx.Done()
		fmt.Printf("Server killed\n")
		server.srv.Shutdown(context.Background())
	}()

	go func() {
		log.Fatal(server.srv.ListenAndServe())
	}()
}
