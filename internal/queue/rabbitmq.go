package queue

import (
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

func NewRabbitMQConnection() (*amqp.Connection, error) {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	return amqp.Dial(url)
}
