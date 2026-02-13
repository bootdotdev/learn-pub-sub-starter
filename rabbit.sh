#!/bin/bash

start_or_run () {
    docker inspect peril_rabbitmq > /dev/null 2>&1

    if [ $? -eq 0 ]; then
        echo "Starting Peril RabbitMQ container..."
        docker start peril_rabbitmq
    else
        echo "Peril RabbitMQ container not found, creating a new one..."
        docker run -d --name peril_rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.13-management
    fi
}

case "$1" in
    start)
        start_or_run
        ;;
    stop)
        echo "Stopping Peril RabbitMQ container..."
        docker stop peril_rabbitmq
        ;;
    logs)
        echo "Fetching logs for Peril RabbitMQ container..."
        docker logs -f peril_rabbitmq
        ;;
    *)
        echo "Usage: $0 {start|stop|logs}"
        exit 1
esac
