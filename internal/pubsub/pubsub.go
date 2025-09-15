package pubsub

import (
    "context"
    "encoding/json"

    amqp "github.com/rabbitmq/amqp091-go"
)

// SimpleQueueType is a minimal enum to represent durable vs transient queues.
type SimpleQueueType int

const (
    // Durable represents a durable queue (survives broker restarts, not auto-deleted).
    Durable SimpleQueueType = iota
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
    queueType SimpleQueueType,
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

// SubscribeJSON declares/binds the queue and starts consuming JSON messages into T.
// It spawns a goroutine that unmarshals, invokes the handler, and acks each message.
func SubscribeJSON[T any](
    conn *amqp.Connection,
    exchange,
    queueName,
    key string,
    queueType SimpleQueueType,
    handler func(T),
) error {
    ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
    if err != nil {
        return err
    }

    // Note: caller is not closing channel here; it must remain open for consumption.
    deliveries, err := ch.Consume(
        q.Name, // queue
        "",     // consumer (auto-generated)
        false,   // auto-ack (false; we manually ack after handling)
        false,   // exclusive
        false,   // no-local
        false,   // no-wait
        nil,     // args
    )
    if err != nil {
        _ = ch.Close()
        return err
    }

    go func() {
        for d := range deliveries {
            var msg T
            if err := json.Unmarshal(d.Body, &msg); err == nil {
                handler(msg)
                _ = d.Ack(false)
            } else {
                // On unmarshal error, reject without requeue to avoid poison messages.
                _ = d.Nack(false, false)
            }
        }
    }()

    return nil
}
