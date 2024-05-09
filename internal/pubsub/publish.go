package pubsub

import (
	"context"
	"encoding/json"

	amqb "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqb.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqb.Publishing{
		ContentType: "application/json",
		Body:        data,
	})
}
