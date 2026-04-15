package api

type CharacterCreateBody struct {
	FirstName string `validate:"required,alpha"`
	LastName  string `validate:"required,alpha"`
	RaceID    uint   `validate:"required,number"`
	Gender    int    `validate:"required"`
}

type CharacterBasicResponse struct {
	ID           uint
	RaceName     string
	FirstName    string
	LastName     string
	LocationName string
	Gender       string
}

type CharacterBasicWithActivityResponse struct {
	CharacterBasicResponse
	CurrentActivityTitle string
}

type CharacterStatResponse struct {
	Code     string
	Name     string
	Value    int
	MaxValue int
}

type CharacterSkillResponse struct {
	Code            string
	Name            string
	IsFightingSkill bool
	Level           int
	CurrentXp       int
	NextLevelXp     int
}

type CharacterInfoResponse struct {
	ID        uint
	FirstName string
	LastName  string
	RaceName  string
	Gender    string
	Power     int
	ScriptID  []byte

	Stats  []CharacterStatResponse
	Skills []CharacterSkillResponse

	Location        LocationResponse
	RespawnLocation LocationResponse
}
