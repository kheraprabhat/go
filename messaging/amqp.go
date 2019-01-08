package messaging

import (
	"fmt"

	"github.com/streadway/amqp"
)

// AMQP is an AMQP structure
type AMQP struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

// Connect is an AMQP connection convenience function
func (_amqp *AMQP) Connect(url string) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	_amqp.Connection = conn

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	_amqp.Channel = ch

	return nil
}

// SendMessage is an AMQP convenience method to send a message to a given queue name
func (_amqp *AMQP) SendMessage(queue string, message Message) error {

	q, err := _amqp.Channel.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		return fmt.Errorf("Failed to declare queue %s: %s", queue, err.Error())
	}

	err = _amqp.Channel.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			Body:          message.Body,
			CorrelationId: message.CorrelationID,
		},
	)
	if err != nil {
		return fmt.Errorf("Failed to send message: %s", err.Error())
	}

	return nil
}

// DeleteMessage is an AMQP convenience method which does nothing, as AMQP does not support message deletion
func (_amqp *AMQP) DeleteMessage(id string) error {
	// No body because AMQP does not support message deletion without consumption
	return nil

}

// ReceiveMessages is an AMQP convenience method to receive messages from a given queue
func (_amqp *AMQP) ReceiveMessages(queue string) (<-chan Message, error) {

	// messages := make([]Message, 0)

	output, err := _amqp.Channel.Consume(
		queue, // queue
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	if err != nil {
		return nil, err
	}

	ret := make(chan Message)
	go func() {
		for message := range output {
			ret <- Message{message.MessageId, message.CorrelationId, message.Body}
		}
	}()

	return ret, nil
}
