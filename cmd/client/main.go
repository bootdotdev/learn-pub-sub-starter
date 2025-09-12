package main

import (
    "fmt"

    amqp "github.com/rabbitmq/amqp091-go"

    "github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
    "github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
    "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func main() {
    fmt.Println("Starting Peril client...")

    // Prompt for username
    username, err := gamelogic.ClientWelcome()
    if err != nil {
        fmt.Println(err)
        return
    }

    // Connect to RabbitMQ
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

    // Ensure the direct exchange exists
    exchCh, err := conn.Channel()
    if err != nil {
        fmt.Println("Failed to open channel to declare exchange:", err)
        return
    }
    if err := pubsub.DeclareDirectExchange(exchCh, routing.ExchangePerilDirect); err != nil {
        _ = exchCh.Close()
        fmt.Println("Failed to declare exchange:", err)
        return
    }
    _ = exchCh.Close()

    // Declare and bind a transient, exclusive queue for this user
    queueName := fmt.Sprintf("%s.%s", routing.PauseKey, username)
    ch, q, err := pubsub.DeclareAndBind(
        conn,
        routing.ExchangePerilDirect, // exchange
        queueName,                   // queueName: pause.username
        routing.PauseKey,            // routing key
        pubsub.Transient,            // queueType
    )
    if err != nil {
        fmt.Println("Failed to declare/bind queue:", err)
        return
    }
    defer func() { _ = ch.Close() }()

    fmt.Printf("Declared transient queue %q bound to exchange %q with key %q\n", q.Name, routing.ExchangePerilDirect, routing.PauseKey)

    // Initialize local game state
    gs := gamelogic.NewGameState(username)

    // Client REPL loop
    for {
        words := gamelogic.GetInput()
        if len(words) == 0 {
            continue
        }
        cmd := words[0]
        switch cmd {
        case "spawn":
            if err := gs.CommandSpawn(words); err != nil {
                fmt.Println(err)
            }
        case "move":
            if _, err := gs.CommandMove(words); err != nil {
                fmt.Println(err)
            } else {
                fmt.Println("Move successful.")
            }
        case "status":
            gs.CommandStatus()
        case "help":
            gamelogic.PrintClientHelp()
        case "spam":
            fmt.Println("Spamming not allowed yet!")
        case "quit":
            gamelogic.PrintQuit()
            return
        default:
            fmt.Println("Unrecognized command. Type 'help' for options.")
        }
    }
}
