package queue

import (
	"context"
	"log"

	"github.com/awesome-academy/golang_baoan_thao/internal/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageHandler func(ctx context.Context, body []byte) error

type RabbitMQConsumer struct {
	conn *amqp.Connection
}

func NewRabbitMQConsumer(conn *amqp.Connection) *RabbitMQConsumer {
	return &RabbitMQConsumer{conn: conn}
}

func (c *RabbitMQConsumer) ConsumeApplicationStatusChanged(
	ctx context.Context,
	queueName string,
	handler MessageHandler,
) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}

	if err := ch.ExchangeDeclare(
		ApplicationExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		_ = ch.Close()
		return err
	}

	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		return err
	}

	if err := ch.QueueBind(
		q.Name,
		events.ApplicationStatusChanged,
		ApplicationExchange,
		false,
		nil,
	); err != nil {
		_ = ch.Close()
		return err
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		return err
	}

	go func() {
		defer ch.Close()

		for {
			select {
			case <-ctx.Done():
				return

			case msg, ok := <-msgs:
				if !ok {
					return
				}

				if err := handler(ctx, msg.Body); err != nil {
					_ = msg.Nack(false, true)
					log.Printf("failed to handle message: %v", err)
					continue
				}

				_ = msg.Ack(false)

			}
		}
	}()

	return nil
}
