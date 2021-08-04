package amqpclient

import (
	"github.com/streadway/amqp"

	"log"
	"time"

	"github.com/pkg/errors"
)

// Broker represents AMQP broker which store connection and connected exchanges
type Broker struct {
	amqpURL     string
	connection  *amqp.Connection
	channel     *amqp.Channel
	exchange    string
	key         string
	undelivered chan amqp.Return
	closed      chan *amqp.Error
	canceled    chan string
}

// NewLogger creates new Broker with new connection
func NewAMQPClient(amqpURL string, exchange string, routingKey string) (*Broker, error) {
	b := &Broker{
		amqpURL:  amqpURL,
		exchange: exchange,
		key:      routingKey,
	}
	err := b.connect()
	if err != nil {
		return b, err
	}
	go b.closedWatcher()
	return b, nil
}

// connect do dial and create channel for work with broker
func (b *Broker) connect() error {
	var err error
	b.connection, err = amqp.Dial(b.amqpURL)
	if err != nil {
		return errors.Wrap(err, "Unable to connect to amqp broker")
	}

	b.channel, err = b.connection.Channel()
	if err != nil {
		return errors.Wrap(err, "Unable to open channel to amqp broker")
	}

	b.closed = make(chan *amqp.Error)
	b.canceled = make(chan string)
	b.undelivered = make(chan amqp.Return)
	b.channel.NotifyClose(b.closed)
	b.channel.NotifyCancel(b.canceled)
	b.channel.NotifyReturn(b.undelivered)
	return nil
}

func (b *Broker) closedWatcher() {
	for {
		<-b.closed
		log.Println("Received closed signal, need to reconnect")

		if err := b.connect(); err != nil {
			log.Printf("Unable to connect to rabbit due to '%s'", err.Error())
		}

		time.Sleep(time.Duration(5) * time.Second)
	}
}

// Close close all channels and connection
func (b *Broker) Close() error {
	return b.connection.Close()
}

// DeliverMessages construct amqp.Publishings from array of array of bytes,
// and publish each one by one to exchange
func (b *Broker) DeliverMessages(data []string) error {
	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
	}
	var err error
	for _, message := range data {
		msg.Body = []byte(message)
		msg.Timestamp = time.Now()
		err = b.channel.Publish(b.exchange, b.key, false, false, msg)
		if err != nil {
			return err
		}
	}
	return nil
}

// Consume returns chan amqp.Delivery for queue
func (b *Broker) Consume(queue string) (<-chan amqp.Delivery, error) {
	return b.channel.Consume(queue, "loggo", false, false, false, false, nil)
}

// CreateExchange create direct durable exchange with inited parameters in constructor
func (b *Broker) CreateExchange() error {
	return b.channel.ExchangeDeclare(b.exchange, "direct", true, false, false, false, nil)
}

// CreateQueue create queue with name and durability parameters
func (b *Broker) CreateQueue(name string, durable bool) (amqp.Queue, error) {
	return b.channel.QueueDeclare(name, durable, false, false, false, nil)
}

// BindQueue binds queue by name to exchange setted in constructor
func (b *Broker) BindQueue(name string) error {
	return b.channel.QueueBind(name, b.key, b.exchange, false, nil)
}
