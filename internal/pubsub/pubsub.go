package pubsub

import (
    "context"
    "encoding/json"

    amqp "github.com/rabbitmq/amqp091-go"
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

