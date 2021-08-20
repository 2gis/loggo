package amqpclient

import (
	"github.com/streadway/amqp"

	"log"
	"time"

	"github.com/pkg/errors"

	"github.com/2gis/loggo/configuration"
)

// AMQPClient which store connection and connected exchanges
type AMQPClient struct {
	url         string
	connection  *amqp.Connection
	channel     *amqp.Channel
	exchange    string
	key         string
	undelivered chan amqp.Return
	closed      chan *amqp.Error
	canceled    chan string
}

// NewAMQPClient creates new AMQPClient with new connection
func NewAMQPClient(config configuration.AMQPTransportConfig) (*AMQPClient, error) {
	b := &AMQPClient{
		url:      config.URL,
		exchange: config.Exchange,
		key:      config.Key,
	}
	err := b.connect()

	if err != nil {
		return b, err
	}

	go b.closedWatcher()
	return b, nil
}

// connect do dial and create channel for work with broker
func (c *AMQPClient) connect() error {
	var err error
	c.connection, err = amqp.Dial(c.url)
	if err != nil {
		return errors.Wrap(err, "Unable to connect to amqp broker")
	}

	c.channel, err = c.connection.Channel()
	if err != nil {
		return errors.Wrap(err, "Unable to open channel to amqp broker")
	}

	c.closed = make(chan *amqp.Error)
	c.canceled = make(chan string)
	c.undelivered = make(chan amqp.Return)
	c.channel.NotifyClose(c.closed)
	c.channel.NotifyCancel(c.canceled)
	c.channel.NotifyReturn(c.undelivered)
	return nil
}

func (c *AMQPClient) closedWatcher() {
	for {
		<-c.closed
		log.Println("Received closed signal, need to reconnect")

		if err := c.connect(); err != nil {
			log.Printf("Unable to connect to rabbit due to '%s'", err.Error())
		}

		time.Sleep(time.Duration(5) * time.Second)
	}
}

// Close close all channels and connection
func (c *AMQPClient) Close() error {
	return c.connection.Close()
}

// DeliverMessages construct amqp.Publishings from array of array of bytes,
// and publish each one by one to exchange
func (c *AMQPClient) DeliverMessages(data []string) error {
	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
	}
	var err error
	for _, message := range data {
		msg.Body = []byte(message)
		msg.Timestamp = time.Now()
		err = c.channel.Publish(c.exchange, c.key, false, false, msg)
		if err != nil {
			return err
		}
	}
	return nil
}

// Consume returns chan amqp.Delivery for queue
func (c *AMQPClient) Consume(queue string) (<-chan amqp.Delivery, error) {
	return c.channel.Consume(queue, "loggo", false, false, false, false, nil)
}

// CreateExchange create direct durable exchange with inited parameters in constructor
func (c *AMQPClient) CreateExchange() error {
	return c.channel.ExchangeDeclare(c.exchange, "direct", true, false, false, false, nil)
}

// CreateQueue create queue with name and durability parameters
func (c *AMQPClient) CreateQueue(name string, durable bool) (amqp.Queue, error) {
	return c.channel.QueueDeclare(name, durable, false, false, false, nil)
}

// BindQueue binds queue by name to exchange setted in constructor
func (c *AMQPClient) BindQueue(name string) error {
	return c.channel.QueueBind(name, c.key, c.exchange, false, nil)
}
