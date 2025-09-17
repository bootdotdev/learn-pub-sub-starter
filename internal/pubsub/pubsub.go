package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
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

	if err := DeclareDirectExchange(ch, routing.ExchangePerilDeadLetters); err != nil {
		_ = ch.Close()
		return nil, amqp.Queue{}, err
	}

	durable := queueType == Durable
	autoDelete := queueType == Transient
	exclusive := queueType == Transient
	args := amqp.Table{
		"x-dead-letter-exchange":    routing.ExchangePerilDeadLetters,
		"x-dead-letter-routing-key": routing.QueuePerilDLQ,
	}

	q, err := ch.QueueDeclare(
		queueName,  // name
		durable,    // durable
		autoDelete, // autoDelete
		exclusive,  // exclusive
		false,      // noWait
		args,       // args
	)
	if err != nil {
		if !isPreconditionFailed(err) {
			_ = ch.Close()
			return nil, amqp.Queue{}, err
		}

		_ = ch.Close()
		ch, err = conn.Channel()
		if err != nil {
			return nil, amqp.Queue{}, err
		}

		if err := DeclareDirectExchange(ch, routing.ExchangePerilDeadLetters); err != nil {
			_ = ch.Close()
			return nil, amqp.Queue{}, err
		}

		if _, delErr := ch.QueueDelete(queueName, false, false, false); delErr != nil && !isNotFound(delErr) {
			_ = ch.Close()
			return nil, amqp.Queue{}, delErr
		}

		q, err = ch.QueueDeclare(
			queueName,  // name
			durable,    // durable
			autoDelete, // autoDelete
			exclusive,  // exclusive
			false,      // noWait
			args,       // args
		)
		if err != nil {
			_ = ch.Close()
			return nil, amqp.Queue{}, err
		}
	}

	if err := ensureDeadLetterQueue(ch, queueType); err != nil {
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
	handler func(T) AckType,
) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	// Note: caller is not closing channel here; it must remain open for consumption.
	deliveries, err := ch.Consume(
		q.Name, // queue
		"",     // consumer (auto-generated)
		false,  // auto-ack (false; we manually ack after handling)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		_ = ch.Close()
		return err
	}

	go func() {
		for d := range deliveries {
			var msg T
			if err := json.Unmarshal(d.Body, &msg); err == nil {
				switch handler(msg) {
				case Ack:
					_ = d.Ack(false)
					log.Printf("SubscribeJSON: Acked message from queue %s", q.Name)
				case NackRequeue:
					_ = d.Nack(false, true)
					log.Printf("SubscribeJSON: Nack (requeue) message from queue %s", q.Name)
				case NackDiscard:
					_ = d.Nack(false, false)
					log.Printf("SubscribeJSON: Nack (discard) message from queue %s", q.Name)
				default:
					_ = d.Nack(false, false)
					log.Printf("SubscribeJSON: Unknown ack type, Nack (discard) message from queue %s", q.Name)
				}
			} else {
				// On unmarshal error, reject without requeue to avoid poison messages.
				_ = d.Nack(false, false)
				log.Printf("SubscribeJSON: Unmarshal error, Nack (discard) message from queue %s: %v", q.Name, err)
			}
		}
	}()

	return nil
}

// AckType determines how a consumed message should be acknowledged.
type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func isPreconditionFailed(err error) bool {
	var amqpErr *amqp.Error
	return errors.As(err, &amqpErr) && amqpErr.Code == 406
}

func isNotFound(err error) bool {
	var amqpErr *amqp.Error
	return errors.As(err, &amqpErr) && amqpErr.Code == 404
}

func ensureDeadLetterQueue(ch *amqp.Channel, queueType SimpleQueueType) error {
	if queueType != Durable {
		return nil
	}

	queueName := routing.QueuePerilDLQ
	routingKey := routing.QueuePerilDLQ

	declare := func() error {
		if _, err := ch.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // autoDelete
			false,     // exclusive
			false,     // noWait
			nil,       // args
		); err != nil {
			return err
		}

		return ch.QueueBind(
			queueName,                        // queue name
			routingKey,                       // routing key
			routing.ExchangePerilDeadLetters, // exchange
			false,                            // noWait
			nil,                              // args
		)
	}

	if err := declare(); err != nil {
		if !isPreconditionFailed(err) {
			return err
		}

		if _, delErr := ch.QueueDelete(queueName, false, false, false); delErr != nil && !isNotFound(delErr) {
			return delErr
		}

		return declare()
	}

	return nil
}
