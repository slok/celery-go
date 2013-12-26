package broker

import ()

// This interface must be implemented by all the backends
type Connection interface {
	Connect() error
	Disconnect() error
}

// has the connection to start the process of getting tasks
type Broker struct {
	Conn Connection
}
