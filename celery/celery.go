package main

import (
	"fmt"
	"os"

	"github.com/slok/celery-go/celery/broker"
)

var ()

const ()

func main() {
	// Load the settings

	// Load the params
	params := map[string]interface{}{
		"something": "something",
	}

	// Prepare stuff
	conn, err := broker.NewAmqpConnection(params)
	if err != nil {
		os.Exit(1)
	}

	broker := broker.Broker{
		Conn: conn,
	}

	broker.Conn.Connect()
	ms, _ := broker.Conn.Consume()

	for m := range ms {
		fmt.Println(m.Body)
	}
}
