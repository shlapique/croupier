package server

// server works with some 'struct Response'
// and with some Preloader[T]

import (
	"context"
	"net/http"
	// "sync"
	// "errors"
	"embed"
	// "fmt"
	"io/fs"
	"log/slog"

	"croupier/internal/downloader"
	"croupier/internal/preloader"
	"croupier/internal/yadisk"
)

type Server[T any] struct {
	l *slog.Logger

	srv     *http.Server
	Backend *Backend[T]
	Port    string
}

type Backend[T any] struct {
	Client     *yadisk.Client
	Preloader  *preloader.Preloader[T]
	Downloader *downloader.Downloader
}

func New[T any](l *slog.Logger, backend *Backend[T], port string, assets embed.FS) *Server[T] {
	if l == nil {
		l = slog.New(slog.Default().Handler())
	}
	l = l.With("pkg", "server")

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
		l:          l,
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
		l:       l,
		srv:     srv,
		Backend: backend,
		Port:    port,
	}
}

func (server *Server[T]) Run(ctx context.Context) {
	server.l.Info("listening", "port", server.Port)

	go func() {
		<-ctx.Done()
		server.l.Info("killed")
		server.srv.Shutdown(context.Background())
	}()

	go func() {
		err := server.srv.ListenAndServe()
		if err != nil {
			server.l.Error("Error occured", "err", err)
			return
		}
	}()
}
