package gamelogic

import (
	"sync"
)

type GameState struct {
	Player Player
	Paused bool
	mu     *sync.RWMutex
}

func NewGameState(username string) *GameState {
	return &GameState{
		Player: Player{
			Username: username,
			Units:    map[int]Unit{},
		},
		Paused: false,
		mu:     &sync.RWMutex{},
	}
}

func (gs *GameState) resumeGame() {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.Paused = false
}

func (gs *GameState) pauseGame() {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.Paused = true
}

func (gs *GameState) isPaused() bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.Paused
}

func (gs *GameState) addUnit(u Unit) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.Player.Units[u.ID] = u
}

func (gs *GameState) removeUnitsInLocation(loc Location) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	for k, v := range gs.Player.Units {
		if v.Location == loc {
			delete(gs.Player.Units, k)
		}
	}
}

func (gs *GameState) UpdateUnit(u Unit) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.Player.Units[u.ID] = u
}

func (gs *GameState) GetUsername() string {
	return gs.Player.Username
}

func (gs *GameState) getUnitsSnap() []Unit {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	Units := []Unit{}
	for _, v := range gs.Player.Units {
		Units = append(Units, v)
	}
	return Units
}

func (gs *GameState) GetUnit(id int) (Unit, bool) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	u, ok := gs.Player.Units[id]
	return u, ok
}

func (gs *GameState) GetPlayerSnap() Player {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	Units := map[int]Unit{}
	for k, v := range gs.Player.Units {
		Units[k] = v
	}
	return Player{
		Username: gs.Player.Username,
		Units:    Units,
	}
}
