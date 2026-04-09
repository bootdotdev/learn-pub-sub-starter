package main

import (
	"fmt"

	"github.com/anand-anshul/peril/internal/gamelogic"
	"github.com/anand-anshul/peril/internal/pubsub"
	"github.com/anand-anshul/peril/internal/routing"
)

func handlerLogs() func(routing.GameLog) pubsub.AckType {
	return func(gl routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")
		err := gamelogic.WriteLog(gl)
		if err != nil {
			// what ack type means "failed, try again later"?
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
