package api

type CharacterCreateBody struct {
	FirstName string `validate:"required,alpha"`
	LastName  string `validate:"required,alpha"`
	RaceID    uint   `validate:"required,number"`
}

type CharacterBasicResponse struct {
	ID           uint
	RaceName     string
	FirstName    string
	LastName     string
	LocationName string
}

type CharacterBasicWithActivityResponse struct {
	CharacterBasicResponse
	CurrentActivityTitle string
}

type CharacterStatResponse struct {
	Code     string
	Value    int
	MaxValue int
}

type CharacterSkillResponse struct {
	Code        string
	Level       int
	CurrentXp   int
	NextLevelXp int
}

type CharacterInfoResponse struct {
	ID        uint
	FirstName string
	LastName  string
	RaceName  string
	Power     int

	Stats  []CharacterStatResponse
	Skills []CharacterSkillResponse

	Location        LocationResponse
	RespawnLocation LocationResponse
}
