package main

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func main() {
	fmt.Println("Starting Peril server...")
	gamelogic.PrintServerHelp()

	connStr := "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(connStr)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ:", err)
		return
	}

	defer func() {
		if conn != nil && !conn.IsClosed() {
			_ = conn.Close()
		}
	}()

	fmt.Println("Successfully connected to RabbitMQ.")

	exchCh, err := conn.Channel()
	if err != nil {
		fmt.Println("Failed to open exchange channel:", err)
		return
	}
	if err := pubsub.DeclareDirectExchange(exchCh, routing.ExchangePerilDirect); err != nil {
		_ = exchCh.Close()
		fmt.Println("Failed to declare exchange:", err)
		return
	}
	if err := pubsub.DeclareTopicExchange(exchCh, routing.ExchangePerilTopic); err != nil {
		_ = exchCh.Close()
		fmt.Println("Failed to declare topic exchange:", err)
		return
	}
	_ = exchCh.Close()

	logKey := fmt.Sprintf("%s.*", routing.GameLogSlug)
	if err := pubsub.SubscribeGob[routing.GameLog](
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		logKey,
		pubsub.Durable,
		handlerGameLog(),
	); err != nil {
		fmt.Println("Failed to subscribe to game logs:", err)
		return
	}

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel:", err)
		return
	}
	defer func() { _ = ch.Close() }()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		cmd := words[0]
		switch cmd {
		case "pause":
			fmt.Println("Sending pause message...")
			msg := routing.PlayingState{IsPaused: true}
			if err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, msg); err != nil {
				fmt.Println("Failed to publish pause message:", err)
			} else {
				fmt.Println("Published pause message.")
			}
		case "resume":
			fmt.Println("Sending resume message...")
			msg := routing.PlayingState{IsPaused: false}
			if err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, msg); err != nil {
				fmt.Println("Failed to publish resume message:", err)
			} else {
				fmt.Println("Published resume message.")
			}
		case "quit":
			fmt.Println("Exiting server REPL...")
			return
		default:
			fmt.Println("Unrecognized command. Type 'help' for options.")
		}
	}
}

func handlerGameLog() func(routing.GameLog) pubsub.AckType {
	return func(gl routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")

		if err := gamelogic.WriteLog(gl); err != nil {
			fmt.Println("Failed to write game log:", err)
			return pubsub.NackRequeue
		}

		return pubsub.Ack
	}
}
