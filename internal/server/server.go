package server
// server works with some 'struct Response'
// and with some Preloader[T]

import (
	"net/http"
	"context"
	// "sync"
	"errors"
	"fmt"
)

type Server struct {
	srv     *http.Server
	Backend *Preloader[T]
	Port    string
}

func New[T any](backend *Preloader[T], port string) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /state", getStateHandler)
	mux.HandleFunc("POST /next", nextHandler)
	mux.HandleFunc("POST /prev", prevHandler)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	return &Server{
		srv: srv,
		Backend: backend,
		Port: port,
	}
}

func (server *Server) Run() {
	log.Printf("server listening on %s\n", server.Port)
	log.Fatal(server.srv.ListenAndServe())
}
