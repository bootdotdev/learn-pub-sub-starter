package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"

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

    // Keep the client running so you can inspect the queue in RabbitMQ UI
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
    fmt.Println("Press Ctrl+C to quit the client...")
    <-sigCh
    fmt.Println("Client shutting down...")
}
