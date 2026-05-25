package main

import (
	"os"
	"fmt"
	"context"
	"os/signal"
	"syscall"
	// "time"

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
	// fmt.Printf("YANDEX_DISK_TOKEN: %s\n", token)

	client := yadisk.New(token)
	resource, err := client.GetMeta(ctx, "kindle/")
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Printf("Resouce: %s\n", resource.Name)
	fmt.Println("FULL: ", resource)
}
