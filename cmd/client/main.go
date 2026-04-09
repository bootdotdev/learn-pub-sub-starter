package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"

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

	channel, err := connection.Channel()
	if err != nil {
		log.Fatalf("could not create a new channel: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	gameState := gamelogic.NewGameState(username)
	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gameState),
	)

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix+".*",
		pubsub.Transient,
		handlerMove(gameState, channel),
	)

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.Durable,
		handlerWar(gameState, channel),
	)

loop:
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		first := words[0]

		switch first {
		case "spawn":

			err = gameState.CommandSpawn(words)
			if err != nil {
				log.Println(err)
			}
		case "move":
			mv, err := gameState.CommandMove(words)
			if err == nil {

				log.Printf("move is successful to: %s", mv.ToLocation)
			}
			pubsub.PublishJSON(
				channel,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+username,
				mv,
			)
			log.Println("move was published successfully")

		case "status":
			gameState.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			if len(words) != 2 {
				continue
			}
			second := words[1]
			num, err := strconv.Atoi(second)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			for range num {
				malLog := gamelogic.GetMaliciousLog()

				err := publishGameLog(
					channel,
					username,
					malLog,
				)
				if err != nil {
					log.Println(err)
				}
			}

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
