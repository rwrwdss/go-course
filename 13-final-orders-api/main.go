package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// ctx отменяется по Ctrl+C (SIGINT) или SIGTERM (так Kubernetes/Docker останавливают контейнер).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, ":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
	log.Println("👋 сервер остановлен корректно")
}
