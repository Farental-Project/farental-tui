package api

type CharacterBasicInfoResponse struct {
	ID        uint
	RaceName  string
	FirstName string
	LastName  string
}

type CharacterStatResponse struct {
	Code     string
	Value    int
	MaxValue int
}

type CharacterInfoResponse struct {
	ID        uint
	FirstName string
	LastName  string
	RaceName  string
	Power     int

	Stats []CharacterStatResponse

	Location LocationResponse
}

type CharacterCreateBody struct {
	FirstName string
	LastName  string
	RaceID    uint
}
