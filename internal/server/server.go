package server

// server works with some 'struct Response'
// and with some Preloader[T]

import (
	"context"
	"net/http"
	// "sync"
	// "errors"
	"embed"
	"fmt"
	"io/fs"
	"log"

	"croupier/internal/downloader"
	"croupier/internal/preloader"
	"croupier/internal/yadisk"
)

type Server[T any] struct {
	srv     *http.Server
	Backend *Backend[T]
	Port    string
}

type Backend[T any] struct {
	Client     *yadisk.Client
	Preloader  *preloader.Preloader[T]
	Downloader *downloader.Downloader
}

func New[T any](backend *Backend[T], port string, assets embed.FS) *Server[T] {
	mux := http.NewServeMux()

	sub, err := fs.Sub(assets, "static")
	if err != nil {
		panic(err)
	}

	mux.Handle("/", http.FileServer(http.FS(sub)))

	ph := &PreloaderHandlers[T]{
		loader: backend.Preloader,
	}

	mux.HandleFunc("GET /state", ph.Current)
	mux.HandleFunc("POST /next", ph.Next)
	mux.HandleFunc("POST /prev", ph.Prev)

	dh := &DownloaderHandlers{
		client:     backend.Client,
		downloader: backend.Downloader,
	}

	mux.HandleFunc("POST /download", dh.Download)
	mux.HandleFunc("POST /cancel", dh.CancelDownloads)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	return &Server[T]{
		srv:     srv,
		Backend: backend,
		Port:    port,
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
