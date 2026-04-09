package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/anand-anshul/peril/internal/gamelogic"
	"github.com/anand-anshul/peril/internal/pubsub"
	"github.com/anand-anshul/peril/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	const rabbitConnString = "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer connection.Close()
	fmt.Println("Peril game server connected to RabbitMQ!")

	channel, err := connection.Channel()
	if err != nil {
		log.Fatalf("could not create a new channel: %v", err)
	}

	err = pubsub.SubscribeGob(
		connection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.Durable,
		handlerLogs(),
	)
	if err != nil {
		log.Fatalf("could not subscribe to logs: %v", err)
	}

	gamelogic.PrintServerHelp()

loop:
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		first := words[0]

		switch first {
		case "pause":
			log.Println("sending a pause message")
			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)
		case "resume":
			log.Println("sending a resume message")
			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)
		case "quit":
			log.Println("exiting")
			break loop
		default:
			log.Println("don't understand the command")
		}

	}

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Println("RabbitMQ connection closed.")
}
