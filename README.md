# EQMRMQ
EQMRMQ is a Go package for simplified interaction with RabbitMQ, specifically designed for sending messages, receiving responses, and consuming messages from queues with ease.

The eqmrmq package utilizes the [github.com/rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go) package, which is an AMQP 0.9.1 Go client library. This library provides the underlying functionality for interacting with RabbitMQ, including features such as establishing connections, creating channels, publishing messages, consuming messages, and handling acknowledgments. By leveraging [github.com/rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go), eqmrmq simplifies the process of sending and receiving messages to and from RabbitMQ queues within Go applications.

### Installation
To install EQMRMQ, use **go get**:
```bash
go get github.com/lacolle87/eqmrmq
```

### Usage
Import EQMRMQ into your Go project:

```go
import "github.com/yourusername/eqmrmq"
```

##### Sending a Message
To send a message to a queue:

```go
// Create a new RabbitMQ connection
conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
if err != nil {
    panic(err)
}
defer conn.Close()

// Create a channel
ch, err := conn.Channel()
if err != nil {
    panic(err)
}
defer ch.Close()

// Send a message
err := eqmrmq.SendMessage(eqmrmq.Message{
    QueueName:     "my_queue",
    Message:       "Hello, RabbitMQ!",
    CorrelationId: "123",
    ReplyQueue:    "",
    Ch:            ch,
})
if err != nil {
    panic(err)
}
```

##### Sending a Message with Response
To send a message to a queue and wait for a response:

```go
// Send a message and wait for response
response, err := eqmrmq.SendToQueueWithResponse("my_queue", "Hello, RabbitMQ!", ch)
if err != nil {
    panic(err)
}
fmt.Println("Response:", string(response))
```
##### Consuming Messages
To consume messages from a queue:

```go
// Define a handler function
handler := func(ch *amqp.Channel, d amqp.Delivery) error {
    fmt.Println("Received message:", string(d.Body))
    return nil
}

// Consume messages
err := eqmrmq.ConsumeMessages(ch, "my_queue", handler)
if err != nil {
    panic(err)
}
```

##### Replying to a Message
To reply to a message:

```go
// Reply to a message
err := eqmrmq.ReplyToMessage(ch, delivery, []byte("Response from server"))
if err != nil {
    panic(err)
}
```

### Acknowledgments
Special thanks to the authors of RabbitMQ and the AMQP 0.9.1 Go client library [github.com/rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go) for providing the underlying functionality used by this package.README
