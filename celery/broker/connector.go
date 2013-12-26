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
	Status      Status
}

// TODO
// Status of the message in the past and for the future
type Status struct {
}
