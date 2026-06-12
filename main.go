package main

import (
	// "bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	// "net/http"
	"log"
	// "log/slog"

	"github.com/google/uuid"

	"croupier/internal/preloader"
	"croupier/internal/server"
	"croupier/internal/yadisk"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	// logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	// logger.Info("started!")

	token := getEnv("YANDEX_DISK_TOKEN", "")
	if token == "" {
		fmt.Println("YANDEX_DISK_TOKEN env var is required!")
		return
	}

	debug := getEnv("DEBUG", "0")
	if debug == "1" {
		fmt.Println("DEBUG mode enabled")
	}

	// BACKEND
	// *******
	// yaclient
	// we need to spec num of items per page
	// => limit
	pageSize := 5
	path := "disk:/kindle/"
	client := yadisk.New(yadisk.Config{
		Token:   token,
		Timeout: 15,
	})
	// get inital info about folder
	meta, err := client.GetMeta(ctx, path, pageSize, 0)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	fmt.Printf("Name: %s\n", meta.Name)
	fmt.Println("Meta: ", meta)
	if meta.Type == "dir" {
		fmt.Println("Its a dir:", meta.Name, "at path:", meta.Path, "!")
		fmt.Println("Embed FULL: ", *meta.Embedded)
	} else {
		fmt.Println("ITS NOT A DIR!")
		return
	}

	total := meta.Embedded.Total
	maxOffset := total / pageSize
	fmt.Printf("total [%d], pageSize [%d], maxOffset [%d]\n", total, pageSize, maxOffset)

	pConfig := preloader.Config[yadisk.Page]{
		Offset:    0,
		MinOffset: 0,
		MaxOffset: maxOffset,
		Size:      5,
		Lag:       2,
		FetchFunc: func(i int) (yadisk.Page, error) {
			if i >= 0 && i <= maxOffset {
				// TODO
				time.Sleep(1)
				resp, err := client.GetMeta(ctx, path, pageSize, i*pageSize)
				// items array
				embArray := &resp.Embedded.Items
				page := yadisk.Page{
					Files: yadisk.MapSubset(embArray, func(r yadisk.Resource) yadisk.File {
						return yadisk.File{
							Id:   uuid.NewString(),
							Name: r.Name,
							Path: r.Path,
							MD5:  r.MD5,
						}
					}),
				}
				return page, err
			} else {
				return yadisk.Page{}, errors.New("i is out of bounds!")
			}
		},
		WorkersNum: 2,
	}

	// create and init preloader
	loader, err := preloader.New[yadisk.Page](ctx, pConfig)
	if err != nil {
		fmt.Println("Unable to create New Loader:", err)
	}
	// init preloader
	loader.Init()

	fmt.Println("Now printing current window state...")
	loader.ShowWindow()

	// create a server
	serv := server.New(loader, "1234")
	serv.Run(ctx)

	log.Println("INFO: waiting for Ctrl+C...")

	<-interrupt
	log.Println("Ctrl+C hit! Shutting down..!")
	cancel()

	//
	// err = loader.LoadRight()
	// err = loader.LoadLeft()
	// loader.ShowWindow()
	// get current PAGE
	// v, err := loader.Sw.GetCell(loader.Lag)
	// firstFile := v.Files[0]
	// resp, err := client.GetDownloadLink(ctx, firstFile)
	// fmt.Printf("LINK TO DOWNLOAD [%s]\n", resp.Href)
	// pth := "./tmp/" + firstFile.Name
	// err = client.DownloadFile(ctx, pth, resp.Href, firstFile.MD5)
}
