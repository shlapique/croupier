package main

import (
	"os"
	"fmt"
	"context"
	"os/signal"
	"syscall"
	"bufio"
	// "time"
	"errors"

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
	
	// const numPages = 15
	// var pageList [numPages]string
	// for i := range pageList {
	// 	pageList[i] = string('A' + rune(i))
	// }

	// fmt.Println("Now printing...")
	// for i, v := range pageList {
	// 	fmt.Println("i:", i, "v:", v)
	// }

	// yaclient
	// we need to spec num of items per page
	// => limit 
	pageSize := 20
	path := "disk:/kindle/"
	client := yadisk.New(yadisk.Config{
		Token: token,
		Timeout: 15,
	})
	meta, err := client.GetMeta(ctx, path, pageSize, 0)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	fmt.Printf("Name: %s\n", meta.Name)
	fmt.Println("Meta: ", meta)
	if meta.Type == "dir" {
		fmt.Println("Its a dir:", meta.Name, "at path:", meta.Path, "!")
		fmt.Println("Embed FULL: ", *meta.Embedded)
	}

	total := meta.Embedded.Total
	maxOffset := total / pageSize
	fmt.Printf("total [%d], pageSize [%d], maxOffset [%d]\n", total, pageSize, maxOffset)

	pConfig := preloader.Config[yadisk.Page]{
		Offset:    0,
		MinOffset: 0,
		MaxOffset: maxOffset,
		Size:      3,
		Lag:       1,
		FetchFunc: func(i int) (yadisk.Page, error) { 
			if i >= 0 && i <= maxOffset { 
				resp, err := client.GetMeta(ctx, path, pageSize, i*pageSize)
				embArray := &resp.Embedded.Items
				page := yadisk.Page{ Names: yadisk.MapNames[yadisk.Resource](embArray, func(r yadisk.Resource) string { return r.Name }) }
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
	loader.Sw.Show()


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
			loader.Sw.Show()
		}
    }
    if err := scanner.Err(); err != nil {
        fmt.Fprintln(os.Stderr, "error:", err)
    }
}
