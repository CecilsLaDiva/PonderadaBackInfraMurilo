package queue

import (
	"context"
	"errors"
	"testing"

	"consumer/internal/store"
)

type mockMessage struct {
	body       []byte
	acked      bool
	nacked     bool
	ackErr     error
	nackErr    error
}

func (m *mockMessage) BodyBytes() []byte { return m.body }
func (m *mockMessage) Ack() error {
	m.acked = true
	return m.ackErr
}
func (m *mockMessage) NackRequeue() error {
	m.nacked = true
	return m.nackErr
}

type mockStore struct {
	insertFn func(context.Context, store.RegistroTelemetria) error
}

func (m mockStore) Insert(ctx context.Context, rec store.RegistroTelemetria) error {
	return m.insertFn(ctx, rec)
}

func (m mockStore) Close() {}

func TestHandleMessageSuccess(t *testing.T) {
	msg := &mockMessage{
		body: []byte(`{"id_dispositivo":"dev-1","instante":"2026-03-22T12:00:00Z","tipo_sensor":"temperatura","natureza_leitura":"analogico","valor_coletado":"25.3"}`),
	}
	s := mockStore{
		insertFn: func(ctx context.Context, rec store.RegistroTelemetria) error {
			if rec.IDDispositivo != "dev-1" {
				t.Fatalf("id_dispositivo inesperado: %s", rec.IDDispositivo)
			}
			return nil
		},
	}

	err := HandleMessage(context.Background(), s, msg)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !msg.acked {
		t.Fatal("expected message to be acked")
	}
	if msg.nacked {
		t.Fatal("expected message not to be nacked")
	}
}

func TestHandleMessageInsertError(t *testing.T) {
	msg := &mockMessage{
		body: []byte(`{"id_dispositivo":"dev-2","instante":"2026-03-22T12:00:00Z","tipo_sensor":"umidade","natureza_leitura":"analogico","valor_coletado":"51.2"}`),
	}
	s := mockStore{
		insertFn: func(ctx context.Context, rec store.RegistroTelemetria) error {
			return errors.New("db failure")
		},
	}

	err := HandleMessage(context.Background(), s, msg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if msg.acked {
		t.Fatal("expected message not to be acked")
	}
	if !msg.nacked {
		t.Fatal("expected message to be nacked")
	}
}
