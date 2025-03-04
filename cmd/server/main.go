package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/rabbitmq/amqp091-go"

)

func main() {
	fmt.Println("Starting Peril server...")
	connectionString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp091.Dial(connectionString)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected successfully")
	channel, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	gamelogic.PrintServerHelp()
	_,_,err = pubsub.DeclareAndBind(conn,routing.ExchangePerilTopic,routing.GameLogSlug,"game_logs.*",1)
	if err != nil {
		log.Fatal(err)
	}
	for{
		inputs:=gamelogic.GetInput()
		if len(inputs)==0 {continue }
		if inputs[0]=="pause"{
			log.Println("Sending a pause message")
			err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
			if err != nil {
				log.Fatal(err)
			}
			continue

		}
		if inputs[0]=="resume"{
			log.Println("Sending a resume message")
			err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
			if err != nil {
				log.Fatal(err)
			}
			continue

		}
		if inputs[0]=="quit"{break}
		log.Println("Invalid command")
	}
	
	
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

}
