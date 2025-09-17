# Changelog

## 2025-09-15 CH4-L1

- internal/pubsub/pubsub.go
  - Added `SubscribeJSON[T any]` to declare/bind a queue, consume deliveries, unmarshal JSON into `T`, invoke the handler, and ack each message (nacks on unmarshal error).
  - Exported queue type: renamed `simpleQueueType` to `SimpleQueueType`; kept constants `Durable` and `Transient` and updated function signatures accordingly.

- cmd/client/main.go
  - Added `handlerPause(gs *gamelogic.GameState) func(routing.PlayingState)`; it defers printing a new prompt and calls `gs.HandlePause` to update local state.
  - Wired pause subscription using `pubsub.SubscribeJSON[routing.PlayingState]` with exchange `routing.ExchangePerilDirect`, queue `pause.<username>`, key `routing.PauseKey`, and `pubsub.Transient` queue type.
  - Removed prior manual queue declare/bind in client; subscription now handles it.

## 2025-09-15 CH4-L4: Assignment

- cmd/client/main.go
  - Subscribed clients to other players' moves on `peril_topic` using queue `army_moves.<username>` bound to key `army_moves.*` (transient) via `pubsub.SubscribeJSON[gamelogic.ArmyMove]`.
  - Added `handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove)`; calls `gs.HandleMove` and defers a new prompt.
  - Published moves after `move` command to `peril_topic` with routing key `army_moves.<username>` using `pubsub.PublishJSON`, logging success on publish.

## 2025-09-15 CH5-L2: Assignment

- internal/pubsub/pubsub.go
  - Updated `SubscribeJSON` handler signature to return an `AckType` (`Ack`, `NackRequeue`, `NackDiscard`).
  - Added `AckType` enum and switched consumer goroutine to Ack/Nack based on handler result.
  - Added log statements on each Ack/Nack (including unmarshal errors which discard).

- cmd/client/main.go
  - Updated `handlerPause` to return `pubsub.AckType` and always `Ack` after handling.
  - Updated `handlerMove` to return `pubsub.AckType`: `Ack` on safe/make-war outcomes; `NackDiscard` on same-player or any other outcome.

## 2025-09-15 CH5-L3: Assignment

- internal/routing/routing.go
  - Added `ExchangePerilDeadLetters` constant for the dead letter exchange name.

- internal/pubsub/pubsub.go
  - Passed an `amqp.Table` with `x-dead-letter-exchange` when declaring queues so RabbitMQ routes failed messages to the dead letter exchange.
  - Ensured the dead letter exchange exists and gracefully recreates queues if they were previously declared without the dead letter configuration.
  - Declared/bound the shared `peril_dlq` queue so failed messages land in the expected place.

## 2025-09-15 CH5-L6: Assignment

- cmd/client/main.go
  - Published `RecognitionOfWar` events to `peril_topic` when moves trigger war, requeuing the move so other clients react.
  - Added a durable war subscription that processes shared war messages with outcome-based Ack/Nack behavior.

## 2025-09-15 CH6-L2: Assignment

- internal/pubsub/pubsub.go
  - Added `PublishGob` helper to send gob-encoded messages with the appropriate content type.

- cmd/client/main.go
  - Logged war outcomes by publishing gob-encoded `GameLog` entries and Ack/Nack based on publish success.
