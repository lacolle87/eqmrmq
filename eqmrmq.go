package eqmrmq

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Message struct {
	QueueName     string
	Msg           []byte
	CorrelationId string
	ReplyQueue    string
	Ch            *amqp.Channel
}

func Connect(rabbitURL string) (*amqp.Connection, error) {
	conn, err := connectRabbitMQ(rabbitURL)
	if err != nil {
		return nil, err
	}
	go monitorConnection(conn, rabbitURL)
	return conn, nil
}

func connectRabbitMQ(rabbitURL string) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error
	baseDelay := 2 * time.Second

	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(rabbitURL)
		if err == nil {
			slog.Info("Connected to RabbitMQ")
			return conn, nil
		}
		slog.Warn(fmt.Sprintf("Failed to connect to RabbitMQ. Retrying in %s...", baseDelay), "error", err)
		time.Sleep(baseDelay)
		baseDelay *= 2
	}
	return nil, fmt.Errorf("failed to connect to RabbitMQ after retries: %w", err)
}

func monitorConnection(conn *amqp.Connection, rabbitURL string) {
	for {
		time.Sleep(30 * time.Second)
		if conn.IsClosed() {
			slog.Warn("RabbitMQ connection lost. Reconnecting...")
			newConn, err := connectRabbitMQ(rabbitURL)
			if err != nil {
				slog.Error("Failed to reconnect to RabbitMQ", "error", err)
				continue
			}
			conn = newConn
		}
	}
}

func (msg Message) Publish() error {
	return msg.Ch.Publish(
		"",
		msg.QueueName,
		false,
		false,
		amqp.Publishing{
			Body:          msg.Msg,
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

func PublishToQueueWithResponse(msg Message) ([]byte, error) {
	var err error
	msg.CorrelationId = GenerateCorrelationId()
	msg.ReplyQueue, err = CreateReplyQueue(msg.Ch)
	if err != nil {
		return nil, err
	}

	if publishErr := msg.Publish(); publishErr != nil {
		return nil, publishErr
	}

	return ReceiveResponse(msg.CorrelationId, msg.ReplyQueue, msg.Ch)
}

func SendMessage(queueName string, message interface{}, ch *amqp.Channel) ([]byte, error) {
	msgJSON, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	msg := Message{
		QueueName: queueName,
		Msg:       msgJSON,
		Ch:        ch,
	}

	return PublishToQueueWithResponse(msg)
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

func ConsumeMessages(ch *amqp.Channel, queueName string, handler func(*amqp.Channel, amqp.Delivery, interface{}) error, arg interface{}) error {
	q, err := DeclareQueue(ch, queueName)
	if err != nil {
		return err
	}

	msgs, errRegisterConsumer := RegisterConsumer(ch, q.Name)
	if errRegisterConsumer != nil {
		return errRegisterConsumer
	}

	for d := range msgs {
		if handlerErr := handler(ch, d, arg); handlerErr != nil {
			_, deleteQueueErr := ch.QueueDelete(d.RoutingKey, false, false, false)
			if deleteQueueErr != nil {
				return fmt.Errorf("failed to delete queue: %w", deleteQueueErr)
			}
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
