package main

import (
	"os"
	"fmt"
	// "os/signal"
	// "time"
)


func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
	}
    return defaultValue
}

func main() {
	token := getEnv("YANDEX_DISK_TOKEN", "NO_TOKEN")

	fmt.Println("TOKEN: ", token)
}
