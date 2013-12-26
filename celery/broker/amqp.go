package broker

import (
	"github.com/streadway/amqp"
)

func NewAmqpConnection(params map[string]string) (Connection, error) {
	return AmqpConnection{}, nil
}

type AmqpConnection struct {
	// Connection params

	// Connection stuff
	conn    *amqp.Connection
	channel *amqp.Channel
}

func (c AmqpConnection) Connect() error {
	return nil
}

func (c AmqpConnection) Disconnect() error {
	return nil
}
