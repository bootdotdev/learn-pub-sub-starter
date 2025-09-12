package pubsub

import (
    "context"
    "encoding/json"

    amqp "github.com/rabbitmq/amqp091-go"
)

// simpleQueueType is a minimal enum to represent durable vs transient queues.
type simpleQueueType int

const (
    // Durable represents a durable queue (survives broker restarts, not auto-deleted).
    Durable simpleQueueType = iota
    // Transient represents a transient, exclusive, auto-deleting queue.
    Transient
)

// PublishJSON publishes a JSON-encoded value to the given exchange and routing key.
func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
    b, err := json.Marshal(val)
    if err != nil {
        return err
    }

    return ch.PublishWithContext(
        context.Background(), // ctx
        exchange,             // exchange
        key,                  // key
        false,                // mandatory
        false,                // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        b,
        },
    )
}

// DeclareDirectExchange ensures a durable direct exchange exists.
func DeclareDirectExchange(ch *amqp.Channel, exchange string) error {
    return ch.ExchangeDeclare(
        exchange, // name
        "direct", // kind
        true,     // durable
        false,    // autoDelete
        false,    // internal
        false,    // noWait
        nil,      // args
    )
}

// DeclareTopicExchange ensures a durable topic exchange exists.
func DeclareTopicExchange(ch *amqp.Channel, exchange string) error {
    return ch.ExchangeDeclare(
        exchange, // name
        "topic",  // kind
        true,     // durable
        false,    // autoDelete
        false,    // internal
        false,    // noWait
        nil,      // args
    )
}

// DeclareAndBind declares a queue (durable or transient) and binds it to an exchange with a routing key.
// It returns the channel and declared queue. The caller is responsible for closing the channel.
func DeclareAndBind(
    conn *amqp.Connection,
    exchange,
    queueName,
    key string,
    queueType simpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
    ch, err := conn.Channel()
    if err != nil {
        return nil, amqp.Queue{}, err
    }

    durable := queueType == Durable
    autoDelete := queueType == Transient
    exclusive := queueType == Transient

    q, err := ch.QueueDeclare(
        queueName, // name
        durable,   // durable
        autoDelete, // autoDelete
        exclusive, // exclusive
        false,     // noWait
        nil,       // args
    )
    if err != nil {
        _ = ch.Close()
        return nil, amqp.Queue{}, err
    }

    if err := ch.QueueBind(
        q.Name,   // name
        key,      // key
        exchange, // exchange
        false,    // noWait
        nil,      // args
    ); err != nil {
        _ = ch.Close()
        return nil, amqp.Queue{}, err
    }

    return ch, q, nil
}
