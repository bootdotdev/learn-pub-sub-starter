# 🛰️ Pub/Sub Game Backend (Peril)

A Go-based event-driven backend demonstrating **Publish/Subscribe architecture** using **RabbitMQ**. Built as part of a real-time strategy game simulation (“Peril”), showcasing scalable and decoupled system design.

## 🚀 Features

* ⚡ Event-driven architecture using Pub/Sub pattern
* 📡 Message broker integration with RabbitMQ (AMQP)
* 🧩 Decoupled services (publisher, subscriber, consumer)
* 🔄 Real-time game events (moves, battles, pause/resume)
* 📦 JSON & Gob-based message serialization
* 🧠 Dead-letter queues & retry handling (Ack/Nack)
* 📈 Scalable with multiple consumers and queues

## 🏗️ Architecture Overview

```
Publisher (Game Server)
        |
        v
   Exchange (RabbitMQ)
        |
  ---------------------
  |        |         |
Queue   Queue     Queue
(Client) (Logs)   (War)
  |        |         |
Consumer Consumer Consumer
```

## 🛠️ Tech Stack

* **Go (Golang)**
* **RabbitMQ**
* **AMQP Protocol**
* **Docker**

## ⚙️ Setup

### 1. Start RabbitMQ

```bash
docker run -it --rm --name rabbitmq \
  -p 5672:5672 -p 15672:15672 \
  rabbitmq:3.13-management
```

Access UI: [http://localhost:15672](http://localhost:15672)
(username: `guest`, password: `guest`)

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Run Server

```bash
go run ./cmd/server
```

### 4. Run Client

```bash
go run ./cmd/client
```

## 🎮 Example Commands (Client)

```bash
spawn europe infantry
move asia 1
status
pause
resume
```

## 📌 Key Concepts

* **Publisher** → Sends events (moves, pause, logs)
* **Exchange** → Routes messages using routing keys
* **Queue** → Stores messages until consumed
* **Consumer** → Processes messages asynchronously

## 🧪 Learning Highlights

* Difference between **sync vs async communication**
* Designing **scalable microservices with Pub/Sub**
* Handling failures with **dead-letter queues**
* Message delivery guarantees (**at-least-once**)


