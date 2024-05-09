package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	connectionStr := "amqp://guest:guest@localhost:5672/"

	connection, err := amqp.Dial(connectionStr)

	defer connection.Close()

	if err != nil {
		panic(err)
	}

	fmt.Println("Connection is successful")

	channel, err := connection.Channel()
	if err != nil {
		log.Fatalf("Cannot create channel %v", err)
	}

	err = pubsub.PublishJSON(
		channel,
		routing.ExchangePerilDirect,
		routing.PauseKey,
		routing.PlayingState{IsPaused: true},
	)
	if err != nil {
		log.Printf("Could not publish time: %v", err)
	}

	fmt.Println("Pause message sent!")
}
