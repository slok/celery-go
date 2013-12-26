package broker

import ()

// This interface must be implemented by all the backends
type Connection interface {
	Connect() error
	Disconnect() error
	Consume() (<-chan *Message, error)
}

// has the connection to start the process of getting tasks
type Broker struct {
	Conn Connection
}

// Base message for all the brokers (Echar broker should translate
// to this object type)
type Message struct {
	ContentType string
	Body        []byte
	// Original message, sometimes neccesary for example ACK, requeue... would
	// need casting (but this will be managed in each broker implementation)
	Original interface{}
}
