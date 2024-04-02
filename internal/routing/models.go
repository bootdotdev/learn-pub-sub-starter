package routing

import "time"

type PlayingState struct {
	IsPaused bool
}

type GameLog struct {
	CurrentTime time.Time
	Message     string
	Username    string
}
