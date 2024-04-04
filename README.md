# learn-pub-sub-starter (Peril)

This is the starter code used in Boot.dev's [Learn Pub/Sub](https://learn.boot.dev/learn-pub-sub) course. It contains:

* The `internal/gamelogic` package, which contains the game logic for the Peril game.
* The `internal/routing` package, which contains some routing constants for the game.
* Stubs of the `cmd/client` and `cmd/server` packages, which are `main` packages that run the client and server for the game.
* The `rabbitmq.sh` script, which starts a RabbitMQ server in a Docker container.

## Running the Game

Make sure you have a RabbitMQ server running locally:

```bash
./rabbit.sh
```

Using separate terminal windows, you can run clients and servers:

```bash
go run ./cmd/server
```

```bash
go run ./cmd/client
```

You will be working in this repository and editing it's files throughout the course. You can peek the solution at any time in the course to make sure you're on the right track.
