package broker

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/streadway/amqp"
)

var (
	defaultFormat = "amqp://%v:%v@%v:%d/"

	// Channel Connection defaults
	defaultUser = "guest"
	defaultPass = "guest"
	defaultHost = "localhost"
	defaultPort = 5672

	// dispatching defauls
	defaultExchangeName     = "celery"
	defaultRoutingKey       = "celery"
	defaultQueueName        = "celery"
	defaultExchangeType     = "direct"
	defaultDurable          = true
	defaultDeleteWhenUnUsed = false
	defaultExclusive        = false
	defaultNoWait           = false
	defaultAutoAck          = false
)

type AmqpConnection struct {
	// settings
	params map[string]interface{}

	// Connection stuff
	conn    *amqp.Connection
	channel *amqp.Channel

	// Use one
	queue    *Queue
	exchange *Exchange
}

func loadSettings(params *map[string]interface{}) {
	// Load default params if not present

	checkSetting := func(key string, dParam interface{}) {
		if _, ok := (*params)[key]; !ok {
			(*params)[key] = dParam
		}
	}

	checkSetting("format", defaultFormat)
	checkSetting("user", defaultUser)
	checkSetting("pass", defaultPass)
	checkSetting("host", defaultHost)
	checkSetting("port", defaultPort)
	checkSetting("exchangeName", defaultExchangeName)
	checkSetting("routingKey", defaultRoutingKey)
	checkSetting("queueName", defaultQueueName)
	checkSetting("exchangeType", defaultExchangeType)
	checkSetting("durable", defaultDurable)
	checkSetting("deleteWhenUnUsed", defaultDeleteWhenUnUsed)
	checkSetting("exclusive", defaultExclusive)
	checkSetting("noWait", defaultNoWait)
	checkSetting("autoAck", defaultAutoAck)

}

// Default constructor
func NewAmqpConnection(params map[string]interface{}) (*AmqpConnection, error) {

	log.Println("New AMQP connection object created")
	a := new(AmqpConnection)

	// Load settings (default ones if neccesary)
	loadSettings(&params)

	a.params = params
	return a, nil
}

// Creates a new amqp connection
func (c *AmqpConnection) Connect() error {
	var err error
	connString := fmt.Sprintf(c.params["format"].(string),
		c.params["user"].(string),
		c.params["pass"].(string),
		c.params["host"].(string),
		c.params["port"].(int))

	c.conn, err = amqp.Dial(connString)
	if err != nil {
		return errors.New("Failed to connect to RabbitMQ")
	}

	c.channel, err = c.createChannel()
	if err != nil {
		return err
	}

	// Set the queue and the exchange
	c.queue = &Queue{
		name:             c.params["queueName"].(string),
		durable:          c.params["durable"].(bool),
		deleteWhenUnUsed: c.params["deleteWhenUnUsed"].(bool),
		exclusive:        c.params["exclusive"].(bool),
		noWait:           c.params["noWait"].(bool),
		params:           nil,
	}

	err = c.declareQueue(c.queue)
	if err != nil {
		return err
	}

	c.exchange = &Exchange{
		name:             c.params["queueName"].(string),
		exchangeType:     c.params["exchangeType"].(string),
		durable:          c.params["durable"].(bool),
		deleteWhenUnUsed: c.params["deleteWhenUnUsed"].(bool),
		exclusive:        c.params["exclusive"].(bool),
		noWait:           c.params["noWait"].(bool),
		params:           nil,
		bindings:         []*Binding{},
	}

	err = c.declareExchange(c.exchange)
	if err != nil {
		return err
	}

	// One exchange to rule them all,
	// One queue to find them,
	// One routing key to bring them all and in the darkness bind them
	bindingName := strings.Join([]string{
		c.params["exchangeName"].(string),
		c.params["routingKey"].(string)},
		"/")
	_, err = c.bind(bindingName, c.params["routingKey"].(string), c.exchange, c.queue)
	if err != nil {
		return err
	}

	return nil
}

// Closes current amqp connection
func (c *AmqpConnection) Disconnect() error {
	err := c.channel.Close()
	if err != nil {
		return errors.New("Failed to close channel")
	}

	err = c.conn.Close()
	if err != nil {
		return errors.New("Failed to close connection")
	}

	log.Println("Connected to Rabbitmq")
	return nil
}

// Start consumming from the connection, returns a channel where the
// messages will be delivered
func (c *AmqpConnection) Consume() (<-chan *Message, error) {
	msgs, err := c.channel.Consume(c.queue.name,
		"",
		c.params["autoAck"].(bool),
		c.params["exclusive"].(bool),
		false,
		c.params["noWait"].(bool),
		nil)

	if err != nil {
		log.Println("Failed consuming queue")
		return nil, errors.New("Failed consuming queue")
	}

	resultChannel := make(chan *Message)

	// this coroutine will get the messages from the amqp and will convert to a
	// custom local format
	log.Println("Start consuming from queue")
	go func() {
		for m := range msgs {
			log.Println("Received package")
			resultChannel <- c.convertMessage(&m)
		}
	}()

	return resultChannel, nil
}

func (c AmqpConnection) createChannel() (*amqp.Channel, error) {

	channel, err := c.conn.Channel()

	if err != nil {
		log.Println("Failed to connect to RabbitMQ")
		return nil, errors.New("Failed to connect to RabbitMQ")
	}
	log.Println("Created channel")
	return channel, nil
}

func (c *AmqpConnection) declareQueue(q *Queue) error {

	// We only need queue information after declaring it
	_, err := c.channel.QueueDeclare(
		q.name,
		q.durable,
		q.deleteWhenUnUsed,
		q.exclusive,
		q.noWait,
		q.params,
	)

	if err != nil {
		log.Println("Failed to declare queue")
		return errors.New("Failed to declare queue")
	}
	log.Println("Declared queue")
	return nil
}

func (c *AmqpConnection) declareExchange(e *Exchange) error {

	// We only need queue information after declaring it
	err := c.channel.ExchangeDeclare(
		e.name,
		e.exchangeType,
		e.durable,
		e.deleteWhenUnUsed,
		e.exclusive,
		e.noWait,
		e.params,
	)

	if err != nil {
		log.Println("Failed to declare exchange")
		return errors.New("Failed to declare exchange")
	}

	log.Println("Declared exchange")
	return nil
}

// Binds the queue to an exchange and creates a Bindind object
func (c *AmqpConnection) bind(name, routingKey string, e *Exchange, q *Queue) (*Binding, error) {
	err := c.channel.QueueBind(q.name, routingKey, e.name, q.noWait, nil)
	if err != nil {
		log.Println("Failed to Bind queue to exchange")
		return nil, errors.New("Failed to Bind queue to exchange")
	}

	b := &Binding{
		name:       name,
		routingKey: routingKey,
		queue:      q,
		exchange:   e,
	}

	// This way we could obtain all the bindings of each exchange
	e.bindings = append(e.bindings, b)

	log.Println("Binded exchange with channel")
	return b, nil

}

// Converts an amqp message to a custom message
func (c *AmqpConnection) convertMessage(m *amqp.Delivery) *Message {
	return &Message{
		ContentType: m.ContentType,
		Body:        m.Body,
		Original:    m,
	}
}

func (c *AmqpConnection) Ack(m *Message) error {
	d := m.Original.(*amqp.Delivery)
	return d.Ack(false)
}

// Aux types
type Queue struct {
	name             string
	durable          bool
	deleteWhenUnUsed bool
	exclusive        bool
	noWait           bool
	params           amqp.Table
}

type Exchange struct {
	name             string
	exchangeType     string
	durable          bool
	deleteWhenUnUsed bool
	exclusive        bool
	noWait           bool
	params           amqp.Table
	bindings         []*Binding //For now use one only
}

type Binding struct {
	name       string
	routingKey string
	queue      *Queue
	exchange   *Exchange
}
