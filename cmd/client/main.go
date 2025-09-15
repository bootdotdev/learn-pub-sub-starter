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

    // Initialize local game state
    gs := gamelogic.NewGameState(username)

    // Subscribe to pause messages for this user
    queueName := fmt.Sprintf("%s.%s", routing.PauseKey, username)
    if err := pubsub.SubscribeJSON[routing.PlayingState](
        conn,
        routing.ExchangePerilDirect,
        queueName,        // pause.username
        routing.PauseKey, // routing key
        pubsub.Transient,
        handlerPause(gs),
    ); err != nil {
        fmt.Println("Failed to subscribe to pause messages:", err)
        return
    }

    // Subscribe to other players' moves using a topic binding
    // Queue: army_moves.username, Key: army_moves.* on peril_topic
    movesQueue := fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username)
    movesKey := fmt.Sprintf("%s.*", routing.ArmyMovesPrefix)
    if err := pubsub.SubscribeJSON[gamelogic.ArmyMove](
        conn,
        routing.ExchangePerilTopic,
        movesQueue,
        movesKey,
        pubsub.Transient,
        handlerMove(gs),
    ); err != nil {
        fmt.Println("Failed to subscribe to army move messages:", err)
        return
    }

    // Create a publishing channel for client-initiated messages
    pubCh, err := conn.Channel()
    if err != nil {
        fmt.Println("Failed to open publishing channel:", err)
        return
    }
    defer func() { _ = pubCh.Close() }()

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
            if mv, err := gs.CommandMove(words); err != nil {
                fmt.Println(err)
            } else {
                // Publish the move to peril_topic with routing key army_moves.username
                rk := fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, gs.GetUsername())
                if err := pubsub.PublishJSON(pubCh, routing.ExchangePerilTopic, rk, mv); err != nil {
                    fmt.Println("Failed to publish move:", err)
                } else {
                    fmt.Println("Move published successfully.")
                }
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

// handlerPause returns a handler that pauses/resumes the local game state.
func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
    return func(ps routing.PlayingState) {
        defer fmt.Print("> ")
        gs.HandlePause(ps)
    }
}

// handlerMove returns a handler that processes incoming ArmyMove messages.
func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
    return func(mv gamelogic.ArmyMove) {
        defer fmt.Print("> ")
        gs.HandleMove(mv)
    }
}
