package eqmrmq

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Message struct {
	QueueName     string
	Message       string
	CorrelationId string
	ReplyQueue    string
	Ch            *amqp.Channel
}

func (msg Message) Publish() error {
	return msg.Ch.Publish(
		"",
		msg.QueueName,
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			Body:          []byte(msg.Message),
			CorrelationId: msg.CorrelationId,
			ReplyTo:       msg.ReplyQueue,
		},
	)
}

func ReceiveResponse(correlationId, replyQueue string, ch *amqp.Channel) ([]byte, error) {
	msgs, err := ch.Consume(
		replyQueue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	for d := range msgs {
		if d.CorrelationId == correlationId {
			_, err = ch.QueueDelete(replyQueue, false, false, false)
			if err != nil {
				return nil, fmt.Errorf("failed to delete reply queue: %w", err)
			}
			return d.Body, nil
		}
	}
	return nil, fmt.Errorf("no response received for correlationId: %s", correlationId)
}

func CreateReplyQueue(channel *amqp.Channel) (string, error) {
	replyQ, err := channel.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return "", err
	}
	return replyQ.Name, nil
}

func GenerateCorrelationId() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func PublishToQueueWithResponse(queueName, message string, ch *amqp.Channel) ([]byte, error) {
	correlationId := GenerateCorrelationId()
	replyQueue, err := CreateReplyQueue(ch)
	if err != nil {
		return nil, err
	}

	msg := Message{
		QueueName:     queueName,
		Message:       message,
		CorrelationId: correlationId,
		ReplyQueue:    replyQueue,
		Ch:            ch,
	}

	if publishErr := msg.Publish(); publishErr != nil {
		return nil, publishErr
	}

	return ReceiveResponse(correlationId, replyQueue, ch)
}

func DeclareQueue(ch *amqp.Channel, queueName string) (amqp.Queue, error) {
	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare %s queue: %w", queueName, err)
	}
	return q, nil
}

func RegisterConsumer(ch *amqp.Channel, queueName string) (<-chan amqp.Delivery, error) {
	msgs, err := ch.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register a consumer for %s: %w", queueName, err)
	}
	return msgs, nil
}

func ConsumeMessages(ch *amqp.Channel, queueName string, handler func(*amqp.Channel, amqp.Delivery) error) error {
	q, err := DeclareQueue(ch, queueName)
	if err != nil {
		return err
	}

	msgs, err := RegisterConsumer(ch, q.Name)
	if err != nil {
		return err
	}

	for d := range msgs {
		if handlerErr := handler(ch, d); handlerErr != nil {
			return fmt.Errorf("handler error: %w", handlerErr)
		}
	}
	return nil
}

func ReplyToMessage(ch *amqp.Channel, d amqp.Delivery, replyData []byte) error {
	return ch.Publish(
		"",
		d.ReplyTo,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Body:          replyData,
		},
	)
}
