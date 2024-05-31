package eqmrmq

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type Message struct {
	QueueName     string
	Message       string
	CorrelationId string
	ReplyQueue    string
	Ch            *amqp.Channel
}

func SendMessage(msgParams Message) error {
	err := msgParams.Ch.Publish(
		"",
		msgParams.QueueName,
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			Body:          []byte(msgParams.Message),
			CorrelationId: msgParams.CorrelationId,
			ReplyTo:       msgParams.ReplyQueue,
		},
	)
	return err
}

func receiveResponse(correlationId, replyQueue string, ch *amqp.Channel) ([]byte, error) {
	fmt.Printf("Waiting for response for correlationId: %s and replyTo: %s\n", correlationId, replyQueue)
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
			fmt.Printf("Response received: %s\n", d.Body)

			_, err = ch.QueueDelete(replyQueue, false, false, false)
			if err != nil {
				return nil, fmt.Errorf("failed to delete reply queue: %w", err)
			}

			return d.Body, nil
		}
	}
	return nil, fmt.Errorf("no response received for correlationId: %s", correlationId)
}

func createReplyQueue(channel *amqp.Channel) (string, error) {
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

func generateCorrelationId() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func SendToQueueWithResponse(queueName, message string, ch *amqp.Channel) ([]byte, error) {
	correlationId := generateCorrelationId()
	replyQueue, err := createReplyQueue(ch)
	if err != nil {
		return nil, err
	}

	msgParams := Message{
		QueueName:     queueName,
		Message:       message,
		CorrelationId: correlationId,
		ReplyQueue:    replyQueue,
		Ch:            ch,
	}

	err = SendMessage(msgParams)
	if err != nil {
		return nil, err
	}
	response, err := receiveResponse(correlationId, replyQueue, ch)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func declareQueue(ch *amqp.Channel, queueName string) (amqp.Queue, error) {
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

func registerConsumer(ch *amqp.Channel, queueName string) (<-chan amqp.Delivery, error) {
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
	q, err := declareQueue(ch, queueName)
	if err != nil {
		return err
	}

	msgs, err := registerConsumer(ch, q.Name)
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
	fmt.Printf("Replying to %s with %s\n", d.ReplyTo, string(replyData))
	err := ch.Publish(
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
	if err != nil {
		return fmt.Errorf("failed to reply to %s: %w", d.ReplyTo, err)
	}
	return nil
}
