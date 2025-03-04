package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	connectionString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp091.Dial(connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}
	queueName := "pause." + userName
	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, queueName, routing.PauseKey, 0)
	if err != nil {
		log.Fatal(err)
	}
	gameState := gamelogic.NewGameState(userName)
	for {
		inputs := gamelogic.GetInput()
		if len(inputs) == 0 {
			continue
		}
		command := inputs[0]
		switch command {
		case "spawn":
			err := gameState.CommandSpawn(inputs)
			if err != nil {
				log.Fatal(err)
			}
		case "move":
			_, err := gameState.CommandMove(inputs)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Move commmand is successful")
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			log.Println("Invalid command")

		}
	}
}
