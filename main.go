package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"croupier/internal/preloader"
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
	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	token := getEnv("YANDEX_DISK_TOKEN", "")
	if token == "" {
		fmt.Println("YANDEX_DISK_TOKEN env var is required!")
		return
	}

	debug := getEnv("DEBUG", "0")
	if debug == "1" {
		fmt.Println("DEBUG mode enabled")
	}

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
		Lag:       4,
		FetchFunc: func(i int) (yadisk.Page, error) {
			if i >= 0 && i <= maxOffset {
				// TODO
				time.Sleep(1)
				resp, err := client.GetMeta(ctx, path, pageSize, i*pageSize)
				// items array
				embArray := &resp.Embedded.Items
				page := yadisk.Page{Files: yadisk.MapSubset(embArray, func(r yadisk.Resource) yadisk.File { return yadisk.File{Name: r.Name, Path: r.Path} })}
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

	fmt.Println("Now printing current window state...")
	loader.ShowWindow()

	// user loop
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter lines (Ctrl+D to end):")
	for scanner.Scan() {
		switch scanner.Text() {
		case "r":
			err = loader.LoadRight()
			if err != nil {
				fmt.Println("EROR in cycle!")
				break
			}
		case "l":
			err = loader.LoadLeft()
			if err != nil {
				fmt.Println("EROR in cycle!")
				break
			}
		case "s":
			loader.ShowWindow()
		// download current pages' [0]th indexed item
		case "d":
			fmt.Printf("Current LAG = %d\n", loader.Lag)
			v, err := loader.Sw.GetCell(loader.Lag)
			if err != nil {
				fmt.Println("EROR download in cycle!")
				break
			}
			if v == nil {
				fmt.Printf("page by lag [%d]: %v\n", loader.Lag, nil)
			} else {
				fmt.Printf("page by lag [%d]: %v\n", loader.Lag, *v)
				firstFile := v.Files[0]
				fmt.Printf("First file: [%v]\n", firstFile)

				fmt.Printf("Lets get FULL DOWNLOAD LINK TO THIS first file!\n")
				resp, err := client.GetDownloadLink(ctx, firstFile.Path)
				if err != nil {
					fmt.Println("Unable to get Donwload link!")
					break
				}
				fmt.Printf("LINK TO DOWNLOAD [%s]\n", resp.Href)
				pth := "./tmp/" + firstFile.Name
				fmt.Printf("Trying to download file to filepath [%s]\n", pth)
				err = client.DownloadFile(ctx, pth, resp.Href)
				if err != nil {
					fmt.Println("Unable to download!")
					break
				}
				fmt.Println("DONE")
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
}
