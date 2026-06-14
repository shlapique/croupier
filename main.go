package main

import (
	// "bufio"
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log"
	"net"
	// "log/slog"

	_ "golang.org/x/crypto/x509roots/fallback"

	"croupier/internal/downloader"
	"croupier/internal/preloader"
	"croupier/internal/server"
	"croupier/internal/yadisk"
)

//go:embed static/*
var staticFS embed.FS

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// id derived from the file's path and MD5
func fileID(path string, md5 *string) string {
	h := sha256.New()
	h.Write([]byte(path))
	h.Write([]byte{0})
	if md5 != nil {
		h.Write([]byte(*md5))
	}
	return hex.EncodeToString(h.Sum(nil))
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

	if dns := os.Getenv("CROUPIER_DNS"); dns != "" {
		net.DefaultResolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
				d := net.Dialer{Timeout: 10 * time.Second}
				return d.DialContext(ctx, "udp", dns)
			},
		}
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

	// create and init preloader
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
							Id:   fileID(r.Path, r.MD5),
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

	loader, err := preloader.New(ctx, pConfig)
	if err != nil {
		fmt.Println("Unable to create New Loader:", err)
	}
	// init preloader
	loader.Init()

	fmt.Println("Now printing current window state...")
	loader.ShowWindow()

	// create downloader
	dConfig := downloader.Config{
		DownloadPath: "./tmp/",
		MaxNumFiles:  50,
		WorkersNum:   2,
	}
	downloader := downloader.New(ctx, dConfig)

	// create a server
	serv := server.New(
		&server.Backend[yadisk.Page]{
			client,
			loader,
			downloader,
		},
		"1234",
		staticFS,
	)
	serv.Run(ctx)

	log.Println("INFO: waiting for Ctrl+C...")

	<-interrupt
	log.Println("Ctrl+C hit! Shutting down..!")
	cancel()

	// giving some time for goroutines to finish
	time.Sleep(time.Second * 2)
}
