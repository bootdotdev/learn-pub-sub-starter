package gamelogic

import (
	"errors"
	"fmt"
)

func (gs *GameState) CommandSpawn(words []string) error {
	if len(words) < 3 {
		return errors.New("usage: spawn <location> <rank>")
	}

	locationName := words[1]
	locations := getAllLocations()
	if _, ok := locations[Location(locationName)]; !ok {
		return fmt.Errorf("error: %s is not a valid location", locationName)
	}

	rank := words[2]
	units := getAllRanks()
	if _, ok := units[UnitRank(rank)]; !ok {
		return fmt.Errorf("error: %s is not a valid unit", rank)
	}

	id := len(gs.getUnitsSnap()) + 1
	gs.addUnit(Unit{
		ID:       id,
		Rank:     UnitRank(rank),
		Location: Location(locationName),
	})

	fmt.Printf("Spawned a(n) %s in %s with id %v\n", rank, locationName, id)
	return nil
}
