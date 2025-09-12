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

    // Declare the connection string
    connStr := "amqp://guest:guest@localhost:5672/"

    // Create a new connection to RabbitMQ
    conn, err := amqp.Dial(connStr)
    if err != nil {
        fmt.Println("Failed to connect to RabbitMQ:", err)
        return
    }

    // Ensure the connection is closed on exit
    defer func() {
        if conn != nil && !conn.IsClosed() {
            _ = conn.Close()
        }
    }()

    fmt.Println("Successfully connected to RabbitMQ.")

    // Create a channel for declaring the exchanges
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
    // Ensure the topic exchange exists
    if err := pubsub.DeclareTopicExchange(exchCh, routing.ExchangePerilTopic); err != nil {
        _ = exchCh.Close()
        fmt.Println("Failed to declare topic exchange:", err)
        return
    }
    _ = exchCh.Close()

    // Declare and bind durable game_logs queue to peril_topic exchange
    logKey := fmt.Sprintf("%s.*", routing.GameLogSlug)
    logCh, _, err := pubsub.DeclareAndBind(
        conn,
        routing.ExchangePerilTopic,
        routing.GameLogSlug,
        logKey,
        pubsub.Durable,
    )
    if err != nil {
        fmt.Println("Failed to declare/bind game_logs queue:", err)
        return
    }
    _ = logCh.Close()

    // Create a channel for publishing
    ch, err := conn.Channel()
    if err != nil {
        fmt.Println("Failed to open a channel:", err)
        return
    }
    defer func() { _ = ch.Close() }()

    // Interactive REPL loop
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
