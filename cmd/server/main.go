package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"

    amqp "github.com/rabbitmq/amqp091-go"

    "github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
    "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func main() {
    fmt.Println("Starting Peril server...")

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

    // Create a channel for publishing
    ch, err := conn.Channel()
    if err != nil {
        fmt.Println("Failed to open a channel:", err)
        return
    }
    defer func() { _ = ch.Close() }()

    // Publish a pause message to the direct exchange
    pause := routing.PlayingState{IsPaused: true}
    if err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, pause); err != nil {
        fmt.Println("Failed to publish pause message:", err)
    } else {
        fmt.Println("Published pause message to exchange", routing.ExchangePerilDirect, "with key", routing.PauseKey)
    }

    // Wait for a signal (e.g., Ctrl+C) to exit
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
    fmt.Println("Press Ctrl+C to stop the server...")

    <-sigCh
    fmt.Println("Shutdown signal received, closing RabbitMQ connection...")

    // Close the connection explicitly on shutdown
    if conn != nil && !conn.IsClosed() {
        _ = conn.Close()
    }
}
