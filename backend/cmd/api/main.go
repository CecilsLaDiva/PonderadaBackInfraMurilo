package main

import (
	"log"
	"os"

	apihttp "backend/internal/http"
	"backend/internal/queue"

	"github.com/gin-gonic/gin"
)

func main() {
	urlRabbit := getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	port := getEnv("PORT", "8080")

	publisher, err := queue.NewRabbitPublisher(urlRabbit)
	if err != nil {
		log.Fatalf("falha ao iniciar publisher: %v", err)
	}
	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("erro fechando publisher: %v", err)
		}
	}()

	router := gin.Default()
	handler := apihttp.NewHandler(publisher)
	handler.RegisterRoutes(router)

	log.Printf("backend subindo na porta :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("erro no servidor http: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
