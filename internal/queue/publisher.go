package queue

import (
	"context"
	"encoding/json"

	"github.com/awesome-academy/golang_baoan_thao/internal/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

const ApplicationExchange = "application.events"

type RabbitMQPublisher struct {
	conn *amqp.Connection
}

func NewRabbitMQPublisher(conn *amqp.Connection) *RabbitMQPublisher {
	return &RabbitMQPublisher{conn: conn}
}

func (p *RabbitMQPublisher) PublishApplicationSubmitted(ctx context.Context, event events.ApplicationSubmittedEvent) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(
		ApplicationExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(
		ctx,
		ApplicationExchange,
		events.ApplicationSubmitted,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

func (p *RabbitMQPublisher) PublishApplicationStatusChanged(ctx context.Context, event events.ApplicationStatusChangedEvent) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(
		ApplicationExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(
		ctx,
		ApplicationExchange,
		events.ApplicationStatusChanged,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

func (p *RabbitMQPublisher) PublishApplicationDeadlineReminder(ctx context.Context, event events.ApplicationDeadlineReminderEvent) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(
		ApplicationExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(
		ctx,
		ApplicationExchange,
		events.ApplicationDeadlineReminder,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}
