package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

var (
	uri          = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	exchangeName = flag.String("exchange", "test-exchange", "Durable AMQP exchange name")
	queueName    = flag.String("queuename", "test-queue", "Durable AMQP exchange name")
	exchangeType = flag.String("exchange-type", "direct", "Exchange type - direct|fanout|topic|x-custom")
	routingKey   = flag.String("key", "test-key", "AMQP routing key")
)

func init() {
	flag.Parse()
}

func main() {
	if err := publish(*uri, *exchangeName, *exchangeType, *routingKey); err != nil {
		log.Fatalf("%s", err)
	}
}

func publish(amqpURI, exchange, exchangeType, routingKey string) error {

	log.Printf("dialing %q", amqpURI)
	connection, err := amqp.Dial(amqpURI)
	if err != nil {
		return fmt.Errorf("Dial: %s", err)
	}
	defer connection.Close()

	log.Printf("got Connection, getting Channel")
	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("Channel: %s", err)
	}

	log.Printf("got Channel, declaring %q Exchange (%q)", exchangeType, exchange)
	if err := channel.ExchangeDeclare(
		exchange,     // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		return fmt.Errorf("Exchange Declare: %s", err)
	}

	q := *queueName
	channel.QueueDeclare(q, true, false, false, false, nil)
	channel.QueueBind(q, routingKey, exchange, false, nil)
	for true {
		time.Sleep(5 * time.Second)
		t := time.Now()
		st := fmt.Sprintf("%d/%d/%d:%d:%d:%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
		body := fmt.Sprintf("%s -> foobar", st)
		fmt.Printf("Sending %s \n",body)
		err = channel.Publish(
			exchange,   // publish to an exchange
			routingKey, // routing to 0 or more queues
			false,      // mandatory
			false,      // immediate
			amqp.Publishing{
				Headers:         amqp.Table{},
				ContentType:     "text/plain",
				ContentEncoding: "",
				Body:            []byte(body),
				DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
				Priority:        0,              // 0-9
				// a bunch of application/implementation-specific fields
			},
		)
		if err != nil {
			panic(fmt.Errorf("Exchange Publish: %s", err))
		}
	}

	return nil
}

