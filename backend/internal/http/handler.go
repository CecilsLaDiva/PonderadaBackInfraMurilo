package http

import (
	"context"
	"net/http"
	"time"

	"backend/internal/model"
	"backend/internal/queue"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	publisher queue.Publisher
}

func NewHandler(publisher queue.Publisher) *Handler {
	return &Handler{publisher: publisher}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.POST("/telemetria", h.PostarTelemetria)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"situacao": "ok"})
	})
}

func (h *Handler) PostarTelemetria(c *gin.Context) {
	var pacote model.PacoteTelemetria
	if err := c.ShouldBindJSON(&pacote); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "json invalido"})
		return
	}

	if err := pacote.Validar(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	// timeout curtinho aqui so pra nao travar req mt tempo
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.publisher.Publish(ctx, pacote); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"erro": "falha ao publicar na fila"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"situacao": "enfileirado"})
}
