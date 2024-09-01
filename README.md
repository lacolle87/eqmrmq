# EQMRMQ

EQMRMQ is a Go package for simplified interaction with RabbitMQ, specifically designed for sending messages, receiving responses, and consuming messages from queues with ease.

The eqmrmq package utilizes the [github.com/rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go) package, which is an AMQP 0.9.1 Go client library. This library provides the underlying functionality for interacting with RabbitMQ, including features such as establishing connections, creating channels, publishing messages, consuming messages, and handling acknowledgments. By leveraging [amqp091-go](https://github.com/rabbitmq/amqp091-go), eqmrmq simplifies the process of sending and receiving messages to and from RabbitMQ queues within Go applications. EQMRMQ utilizes the `slog` package for structured logging.

### Installation

To install EQMRMQ, use **go get**:

```bash
go get github.com/lacolle87/eqmrmq
```

### Usage

Import EQMRMQ into your Go project:

```go
import "github.com/lacolle87/eqmrmq"
```

### Establishing a Connection

To establish a connection to RabbitMQ and monitor it:

```go
package main

import (
	"log"
	"github.com/lacolle87/eqmrmq"
)

func main() {
	rabbitURL := "amqp://guest:guest@localhost:5672/"
	conn, err := eqmrmq.Connect(rabbitURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
}
```

The `Connect` function establishes a connection to RabbitMQ and starts monitoring it to automatically reconnect if the connection is lost.

### Publishing a Message

To publish a message to a queue:

```go
package main

import (
	"log"
	"github.com/lacolle87/eqmrmq"
)

func main() {
    // Create a new RabbitMQ connection
    rabbitURL := "amqp://guest:guest@localhost:5672/"
    conn, err := eqmrmq.Connect(rabbitURL)
    if err != nil {
        log.Fatalf("Failed to connect to RabbitMQ: %v", err)
    }
    defer conn.Close()

    // Create a channel
    ch, err := conn.Channel()
    if err != nil {
        panic(err)
    }
    defer ch.Close()

    // Generate a correlation ID
    correlationId := eqmrmq.GenerateCorrelationId()

    // Send a message
    msg := eqmrmq.Message{
        QueueName:     "my_queue",
        Message:       []byte("Hello, RabbitMQ!"),  // Message is now a byte slice
        CorrelationId: correlationId,
        ReplyQueue:    "", // Assuming no reply queue is needed here
        Ch:            ch,
    }

    err = msg.Publish()
    if err != nil {
        log.Fatalf("Failed to publish message: %v", err)
    }
}
```

### Publishing a Message with Response

To send a message to a queue and wait for a response:

```go
package main

import (
	"fmt"
	"log"
	"github.com/lacolle87/eqmrmq"
)

func main() {
    rabbitURL := "amqp://guest:guest@localhost:5672/"
    conn, err := eqmrmq.Connect(rabbitURL)
    if err != nil {
        log.Fatalf("Failed to connect to RabbitMQ: %v", err)
    }
    defer conn.Close()

    ch, err := conn.Channel()
    if err != nil {
        panic(err)
    }
    defer ch.Close()

    // Send a message and wait for response
    msg := eqmrmq.Message{
        QueueName:   "my_queue",
        Message:     []byte("Hello, RabbitMQ!"),
        Ch:          ch,
    }
    response, err := eqmrmq.PublishToQueueWithResponse(msg)
    if err != nil {
        panic(err)
    }
    fmt.Println("Response:", string(response))
}
```

### Consuming Messages

To consume messages from a queue:

```go
package main

import (
	"fmt"
	"log"
	"github.com/lacolle87/eqmrmq"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
    rabbitURL := "amqp://guest:guest@localhost:5672/"
    conn, err := eqmrmq.Connect(rabbitURL)
    if err != nil {
        log.Fatalf("Failed to connect to RabbitMQ: %v", err)
    }
    defer conn.Close()

    ch, err := conn.Channel()
    if err != nil {
        panic(err)
    }
    defer ch.Close()

    // Define a handler function
    handler := func(ch *amqp.Channel, d amqp.Delivery, arg interface{}) error {
        fmt.Println("Received message:", string(d.Body))
        return nil
    }

    // Any additional argument you want to pass to the handler
    additionalArg := "my_custom_argument" // This can be any type

    // Consume messages
    err = eqmrmq.ConsumeMessages(ch, "my_queue", handler, additionalArg)
    if err != nil {
        panic(err)
    }
}
```

### Replying to a Message

To reply to a message:

```go
package main

import (
	"github.com/lacolle87/eqmrmq"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
    rabbitURL := "amqp://guest:guest@localhost:5672/"
    conn, err := eqmrmq.Connect(rabbitURL)
    if err != nil {
        log.Fatalf("Failed to connect to RabbitMQ: %v", err)
    }
    defer conn.Close()

    ch, err := conn.Channel()
    if err != nil {
        panic(err)
    }
    defer ch.Close()

    // Assuming 'delivery' is an amqp.Delivery received in a consumer
    var delivery amqp.Delivery // This would be your actual delivery object

    // Reply to a message
    err = eqmrmq.ReplyToMessage(ch, delivery, []byte("Response from server"))
    if err != nil {
        panic(err)
    }
}
```

### Acknowledgments

Special thanks to the authors of RabbitMQ and the AMQP 0.9.1 Go client library [github.com/rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go) for providing the underlying functionality used by this package!

---

This README is now aligned with your updated codebase, where `ContentType` is no longer required.