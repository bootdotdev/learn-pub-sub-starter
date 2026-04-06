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
	fmt.Println("Starting Peril client...")

	const rabbitConnString = "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer connection.Close()
	fmt.Println("Peril game client connected to RabbitMQ!")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.Transient,
	)

	gameState := gamelogic.NewGameState(username)

loop:
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		first := words[0]

		switch first {
		case "spawn":
			err := gameState.CommandSpawn(words)
			if err != nil {
				log.Println(err)
			}
		case "move":
			mv, err := gameState.CommandMove(words)
			if err == nil {

				log.Printf("move is successful to: %s", mv.ToLocation)
			}

		case "status":
			gameState.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			fmt.Println("Spamming not allowed yet!")

		case "quit":
			gamelogic.PrintQuit()
			break loop
		default:
			log.Println("don't understand the command")
			continue
		}

	}

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

}
