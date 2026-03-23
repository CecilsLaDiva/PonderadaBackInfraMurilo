package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/model"

	"github.com/gin-gonic/gin"
)

type mockPublisher struct {
	publishFn func(context.Context, model.PacoteTelemetria) error
}

func (m mockPublisher) Publish(ctx context.Context, packet model.PacoteTelemetria) error {
	return m.publishFn(ctx, packet)
}

func (m mockPublisher) Close() error { return nil }

func TestPostTelemetryAccepted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pub := mockPublisher{
		publishFn: func(ctx context.Context, packet model.PacoteTelemetria) error {
			if packet.IDDispositivo != "dev-1" {
				t.Fatalf("id_dispositivo inesperado: %s", packet.IDDispositivo)
			}
			return nil
		},
	}

	handler := NewHandler(pub)
	router := gin.New()
	handler.RegisterRoutes(router)

	payload := map[string]string{
		"id_dispositivo":  "dev-1",
		"instante":        "2026-03-22T12:00:00Z",
		"tipo_sensor":     "temperatura",
		"natureza_leitura":"analogico",
		"valor_coletado":  "25.3",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/telemetria", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, w.Code)
	}
}

func TestPostTelemetryErroValidacao(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pub := mockPublisher{publishFn: func(ctx context.Context, packet model.PacoteTelemetria) error { return nil }}
	handler := NewHandler(pub)
	router := gin.New()
	handler.RegisterRoutes(router)

	payload := map[string]string{
		"id_dispositivo":  "dev-1",
		"instante":        "data-bugada",
		"tipo_sensor":     "temperatura",
		"natureza_leitura":"analogico",
		"valor_coletado":  "25.3",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/telemetria", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
