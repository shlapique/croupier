package server

import (
	"net/http"
	"context"
	// "sync"
	"errors"
	"fmt"
)

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
