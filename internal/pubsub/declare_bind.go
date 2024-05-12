package pubsub

import (
	amqb "github.com/rabbitmq/amqp091-go"
)

func DeclareAndBind(
	conn *amqb.Connection,
	exchange, queueName, key string,
	simpleQueueType int,
) (*amqb.Channel, amqb.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, amqb.Queue{}, err
	}

	queue, err := channel.QueueDeclare(queueName, false, true, true, false, nil)
	if err != nil {
		return nil, amqb.Queue{}, err
	}

	if err = channel.QueueBind(queueName, key, exchange, false, nil); err != nil {
		return nil, amqb.Queue{}, nil
	}

	return channel, queue, nil
}
