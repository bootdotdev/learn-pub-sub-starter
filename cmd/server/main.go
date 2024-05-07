package main

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	connectionStr := "amqp://guest:guest@localhost:5672/"

	connection, err := amqp.Dial(connectionStr)
	if err != nil {
		panic(err)
	}

	defer connection.Close()
}
