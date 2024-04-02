package gamelogic

import (
	"fmt"
)

type WarOutcome int

const (
	WarOutcomeNotInvolved WarOutcome = iota
	WarOutcomeNoUnits
	WarOutcomeYouWon
	WarOutcomeOpponentWon
	WarOutcomeDraw
)

func (gs *GameState) HandleWar(rw RecognitionOfWar) (outcome WarOutcome, winner string, loser string) {
	defer fmt.Println("------------------------")
	fmt.Println()
	fmt.Println("==== War Declared ====")
	fmt.Printf("%s has declared war on %s!\n", rw.Attacker.Username, rw.Defender.Username)

	player := gs.GetPlayerSnap()

	if player.Username == rw.Defender.Username {
		fmt.Printf("%s, you published the war.\n", player.Username)
		return WarOutcomeNotInvolved, "", ""
	}

	if player.Username != rw.Attacker.Username {
		fmt.Printf("%s, you are not involved in this war.\n", player.Username)
		return WarOutcomeNotInvolved, "", ""
	}

	overlappingLocation := getOverlappingLocation(rw.Attacker, rw.Defender)
	if overlappingLocation == "" {
		fmt.Printf("Error! No units are in the same location. No war will be fought.\n")
		return WarOutcomeNoUnits, "", ""
	}

	attackerUnits := []Unit{}
	defenderUnits := []Unit{}
	for _, unit := range rw.Attacker.Units {
		if unit.Location == overlappingLocation {
			attackerUnits = append(attackerUnits, unit)
		}
	}
	for _, unit := range rw.Defender.Units {
		if unit.Location == overlappingLocation {
			defenderUnits = append(defenderUnits, unit)
		}
	}

	fmt.Printf("%s's units:\n", rw.Attacker.Username)
	for _, unit := range attackerUnits {
		fmt.Printf("  * %v\n", unit.Rank)
	}
	fmt.Printf("%s's units:\n", rw.Defender.Username)
	for _, unit := range defenderUnits {
		fmt.Printf("  * %v\n", unit.Rank)
	}
	attackerPower := unitsToPowerLevel(attackerUnits)
	defenderPower := unitsToPowerLevel(defenderUnits)
	fmt.Printf("Attacker has a power level of %v\n", attackerPower)
	fmt.Printf("Defender has a power level of %v\n", defenderPower)
	if attackerPower > defenderPower {
		fmt.Printf("%s has won the war!\n", rw.Attacker.Username)
		if player.Username == rw.Defender.Username {
			fmt.Println("You have lost the war!")
			gs.removeUnitsInLocation(overlappingLocation)
			fmt.Printf("Your units in %s have been killed.\n", overlappingLocation)
			return WarOutcomeOpponentWon, rw.Attacker.Username, rw.Defender.Username
		}
		return WarOutcomeYouWon, rw.Attacker.Username, rw.Defender.Username
	} else if defenderPower > attackerPower {
		fmt.Printf("%s has won the war!\n", rw.Defender.Username)
		if player.Username == rw.Attacker.Username {
			fmt.Println("You have lost the war!")
			gs.removeUnitsInLocation(overlappingLocation)
			fmt.Printf("Your units in %s have been killed.\n", overlappingLocation)
			return WarOutcomeOpponentWon, rw.Defender.Username, rw.Attacker.Username
		}
		return WarOutcomeYouWon, rw.Defender.Username, rw.Attacker.Username
	}
	fmt.Println("The war ended in a draw!")
	fmt.Printf("Your units in %s have been killed.\n", overlappingLocation)
	gs.removeUnitsInLocation(overlappingLocation)
	return WarOutcomeDraw, rw.Attacker.Username, rw.Defender.Username
}

func unitsToPowerLevel(units []Unit) int {
	power := 0
	for _, unit := range units {
		if unit.Rank == RankArtillery {
			power += 10
		}
		if unit.Rank == RankCavalry {
			power += 5
		}
		if unit.Rank == RankInfantry {
			power += 1
		}
	}
	return power
}
