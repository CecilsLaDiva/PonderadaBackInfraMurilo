package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend/internal/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

const NomeFila = "fila_telemetria"

type Publisher interface {
	Publish(context.Context, model.PacoteTelemetria) error
	Close() error
}

type RabbitPublisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewRabbitPublisher(amqpURL string) (*RabbitPublisher, error) {
	var conn *amqp.Connection
	var err error

	// fiz um retry simples aqui so para subir junto no compose
	for i := 0; i < 12; i++ {
		conn, err = amqp.Dial(amqpURL)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar no rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("erro ao abrir canal rabbitmq: %w", err)
	}

	_, err = ch.QueueDeclare(
		NomeFila,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("erro ao declarar fila: %w", err)
	}

	return &RabbitPublisher{conn: conn, ch: ch}, nil
}

func (p *RabbitPublisher) Publish(ctx context.Context, pacote model.PacoteTelemetria) error {
	body, err := json.Marshal(pacote)
	if err != nil {
		return fmt.Errorf("erro ao converter payload pra json: %w", err)
	}

	err = p.ch.PublishWithContext(
		ctx,
		"",
		NomeFila,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now().UTC(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("erro ao publicar na fila: %w", err)
	}

	return nil
}

func (p *RabbitPublisher) Close() error {
	var channelErr error
	if p.ch != nil {
		channelErr = p.ch.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return channelErr
}
