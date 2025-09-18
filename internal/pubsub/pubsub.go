package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
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

// PublishGob publishes a gob-encoded value to the given exchange and routing key.
func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(val); err != nil {
		return err
	}

	return ch.PublishWithContext(
		context.Background(), // ctx
		exchange,             // exchange
		key,                  // key
		false,                // mandatory
		false,                // immediate
		amqp.Publishing{
			ContentType: "application/gob",
			Body:        buf.Bytes(),
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

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
	subscriber string,
) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	if err := ch.Qos(10, 0, false); err != nil {
		_ = ch.Close()
		return err
	}

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
			msg, err := unmarshaller(d.Body)
			if err != nil {
				_ = d.Nack(false, false)
				log.Printf("%s: Unmarshal error, Nack (discard) message from queue %s: %v", subscriber, q.Name, err)
				continue
			}

			switch handler(msg) {
			case Ack:
				_ = d.Ack(false)
				log.Printf("%s: Acked message from queue %s", subscriber, q.Name)
			case NackRequeue:
				_ = d.Nack(false, true)
				log.Printf("%s: Nack (requeue) message from queue %s", subscriber, q.Name)
			case NackDiscard:
				_ = d.Nack(false, false)
				log.Printf("%s: Nack (discard) message from queue %s", subscriber, q.Name)
			default:
				_ = d.Nack(false, false)
				log.Printf("%s: Unknown ack type, Nack (discard) message from queue %s", subscriber, q.Name)
			}
		}
	}()

	return nil
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
	return subscribe(
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		func(body []byte) (T, error) {
			var msg T
			return msg, json.Unmarshal(body, &msg)
		},
		"SubscribeJSON",
	)
}

// SubscribeGob declares/binds the queue and starts consuming gob messages into T.
// It mirrors SubscribeJSON but decodes gob payloads.
func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	return subscribe(
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		func(body []byte) (T, error) {
			var msg T
			dec := gob.NewDecoder(bytes.NewReader(body))
			return msg, dec.Decode(&msg)
		},
		"SubscribeGob",
	)
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
