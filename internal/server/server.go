package server
// server works with some 'struct Response'
// and with some Preloader[T]

import (
	"net/http"
	"context"
	// "sync"
	// "errors"
	"log"
	"fmt"

	"croupier/internal/preloader"
)

type Server[T any] struct {
	srv     *http.Server
	Backend *preloader.Preloader[T]
	Port    string
}

func New[T any](backend *preloader.Preloader[T], port string) *Server[T] {
	mux := http.NewServeMux()

	mux.Handle("GET /state", stateHandler[T]{ loader: backend })
	mux.HandleFunc("POST /next", nextHandler)
	mux.HandleFunc("POST /prev", prevHandler)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	return &Server[T]{
		srv: srv,
		Backend: backend,
		Port: port,
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
