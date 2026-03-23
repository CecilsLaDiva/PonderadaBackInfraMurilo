package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"consumer/internal/queue"
	"consumer/internal/store"
)

func main() {
	urlRabbit := getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	urlBanco := getEnv("DATABASE_URL", "postgres://postgres:postgres@postgres:5432/telemetria?sslmode=disable")

	dbStore, err := store.NewPostgresStore(urlBanco)
	if err != nil {
		log.Fatalf("falha ao iniciar conexao com postgres: %v", err)
	}
	defer dbStore.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := queue.ConsumeForever(ctx, urlRabbit, dbStore); err != nil {
		log.Fatalf("consumidor caiu: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
