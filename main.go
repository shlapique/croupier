package main

import (
	"os"
	"fmt"
	// "context"
	"os/signal"
	"syscall"
	// "time"

	"croupier/internal/yadisk"
)

// type RingBuffer struct {
// 	Buffer []string

// 	Head   *string
// 	Tail   *string
// 	Size   int
// }


func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
	}
    return defaultValue
}

func main() {
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	token := getEnv("YANDEX_DISK_TOKEN", "")
	if token == "" {
		fmt.Println("YANDEX_DISK_TOKEN env var is required!")
		return 
	}
	
	const numPages = 15
	var pageList [numPages]string
	for i := range pageList {
		pageList[i] = string('A' + rune(i))
	}

	fmt.Println("Now printing...")
	for i, v := range pageList {
		fmt.Println("i:", i, "v:", v)
	}

	loader, err := yadisk.NewPreloader[string](numPages, 5, 2)
	if err != nil {
		fmt.Println("error occured:", err)
	}

	for _, v := range pageList {
		err := loader.LoadRight(&v)
		if err != nil {
			fmt.Println("FUKC:", err)
			break
		}
	}

	fmt.Println("Now printing current window state...")
	loader.ShowWindow()

	// fmt.Printf("YANDEX_DISK_TOKEN: %s\n", token)

	// client := yadisk.New(token)
	// meta, err := client.GetMeta(ctx, "disk:/kindle/67.apk")
	// if err != nil {
	// 	fmt.Println("Error: ", err)
	// }

	// fmt.Printf("Name: %s\n", meta.Name)
	// fmt.Println("Meta: ", meta)
	// if meta.Type == "dir" {
	// 	fmt.Println("Its a dir:", meta.Name, "at path:", meta.Path, "!")
	// 	fmt.Println("Embed FULL: ", *meta.Embedded)
	// }
}
