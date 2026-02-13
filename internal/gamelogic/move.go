package gamelogic

import (
	"errors"
	"fmt"
	"strconv"
)

type MoveOutcome int

const (
	MoveOutcomeSamePlayer MoveOutcome = iota
	MoveOutComeSafe
	MoveOutcomeMakeWar
)

func (gs *GameState) HandleMove(move ArmyMove) MoveOutcome {
	defer fmt.Println("------------------------")
	player := gs.GetPlayerSnap()

	fmt.Println()
	fmt.Println("==== Move Detected ====")
	fmt.Printf("%s is moving %v unit(s) to %s\n", move.Player.Username, len(move.Units), move.ToLocation)
	for _, unit := range move.Units {
		fmt.Printf("* %v\n", unit.Rank)
	}

	if player.Username == move.Player.Username {
		return MoveOutcomeSamePlayer
	}

	overlappingLocation := getOverlappingLocation(player, move.Player)
	if overlappingLocation != "" {
		fmt.Printf("You have units in %s! You are at war with %s!\n", overlappingLocation, move.Player.Username)
		return MoveOutcomeMakeWar
	}
	fmt.Printf("You are safe from %s's units.\n", move.Player.Username)
	return MoveOutComeSafe
}

func getOverlappingLocation(p1 Player, p2 Player) Location {
	for _, u1 := range p1.Units {
		for _, u2 := range p2.Units {
			if u1.Location == u2.Location {
				return u1.Location
			}
		}
	}
	return ""
}

func (gs *GameState) CommandMove(words []string) (ArmyMove, error) {
	if gs.isPaused() {
		return ArmyMove{}, errors.New("the game is paused, you can not move units")
	}
	if len(words) < 3 {
		return ArmyMove{}, errors.New("usage: move <location> <unitID> <unitID> <unitID> etc")
	}
	newLocation := Location(words[1])
	locations := getAllLocations()
	if _, ok := locations[newLocation]; !ok {
		return ArmyMove{}, fmt.Errorf("error: %s is not a valid location", newLocation)
	}
	unitIDs := []int{}
	for _, word := range words[2:] {
		id := word
		unitID, err := strconv.Atoi(id)
		if err != nil {
			return ArmyMove{}, fmt.Errorf("error: %s is not a valid unit ID", id)
		}
		unitIDs = append(unitIDs, unitID)
	}

	newUnits := []Unit{}
	for _, unitID := range unitIDs {
		unit, ok := gs.GetUnit(unitID)
		if !ok {
			return ArmyMove{}, fmt.Errorf("error: unit with ID %v not found", unitID)
		}
		unit.Location = newLocation
		gs.UpdateUnit(unit)
		newUnits = append(newUnits, unit)
	}

	mv := ArmyMove{
		ToLocation: newLocation,
		Units:      newUnits,
		Player:     gs.GetPlayerSnap(),
	}
	fmt.Printf("Moved %v units to %s\n", len(mv.Units), mv.ToLocation)
	return mv, nil
}
