package broker

import (
	"github.com/streadway/amqp" // TODO: This shoudn't be here ~~~~~smell
)

// This interface must be implemented by all the backends
type Connection interface {
	Connect() error
	Disconnect() error
	Consume() (<-chan amqp.Delivery, error) // TODO: Create a generic message channel
}

// has the connection to start the process of getting tasks
type Broker struct {
	Conn Connection
}
