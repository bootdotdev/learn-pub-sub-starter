package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"

    amqp "github.com/rabbitmq/amqp091-go"
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
