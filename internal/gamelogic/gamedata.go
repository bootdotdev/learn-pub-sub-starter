package gamelogic

type Player struct {
	Username string
	Units    map[int]Unit
}

type UnitRank string

const (
	RankInfantry  = "infantry"
	RankCavalry   = "cavalry"
	RankArtillery = "artillery"
)

type Unit struct {
	ID       int
	Rank     UnitRank
	Location Location
}

type ArmyMove struct {
	Player     Player
	Units      []Unit
	ToLocation Location
}

type RecognitionOfWar struct {
	Attacker Player
	Defender Player
}

type Location string

func getAllRanks() map[UnitRank]struct{} {
	return map[UnitRank]struct{}{
		RankInfantry:  {},
		RankCavalry:   {},
		RankArtillery: {},
	}
}

func getAllLocations() map[Location]struct{} {
	return map[Location]struct{}{
		"americas":   {},
		"europe":     {},
		"africa":     {},
		"asia":       {},
		"australia":  {},
		"antarctica": {},
	}
}
