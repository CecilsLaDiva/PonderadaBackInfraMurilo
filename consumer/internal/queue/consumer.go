package queue

import (
	"context"
	"fmt"
	"log"
	"time"

	"consumer/internal/store"

	amqp "github.com/rabbitmq/amqp091-go"
)

const NomeFila = "fila_telemetria"

type message interface {
	BodyBytes() []byte
	Ack() error
	NackRequeue() error
}

type amqpMessage struct {
	delivery amqp.Delivery
}

func (m amqpMessage) BodyBytes() []byte {
	return m.delivery.Body
}

func (m amqpMessage) Ack() error {
	return m.delivery.Ack(false)
}

func (m amqpMessage) NackRequeue() error {
	return m.delivery.Nack(false, true)
}

func HandleMessage(ctx context.Context, s store.TelemetryStore, msg message) error {
	record, err := store.DecodificarTelemetria(msg.BodyBytes())
	if err != nil {
		_ = msg.NackRequeue()
		return err
	}

	if err := s.Insert(ctx, record); err != nil {
		_ = msg.NackRequeue()
		return err
	}

	if err := msg.Ack(); err != nil {
		return fmt.Errorf("erro dando ack na msg: %w", err)
	}

	return nil
}

func ConsumeForever(ctx context.Context, amqpURL string, s store.TelemetryStore) error {
	var conn *amqp.Connection
	var err error

	for i := 0; i < 12; i++ {
		conn, err = amqp.Dial(amqpURL)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("erro ao conectar no rabbitmq: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("erro ao abrir canal rabbitmq: %w", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		NomeFila,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("erro ao declarar fila: %w", err)
	}

	deliveries, err := ch.Consume(
		NomeFila,
		"",
		false, // ack na mao dps q gravar no banco
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("erro registrando consumidor: %w", err)
	}

	// bem simples aqui: fica escutando e gravando no banco sem firula
	log.Println("consumidor subiu, esperando msgs...")
	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("canal de entrega fechou")
			}
			if err := HandleMessage(ctx, s, amqpMessage{delivery: delivery}); err != nil {
				log.Printf("deu ruim ao processar msg: %v", err)
			}
		}
	}
}
