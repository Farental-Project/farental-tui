package api

type InfoResponse struct {
	ID          uint
	Code        string
	Name        string
	Description string
}

type RaceResponse struct {
	ID               uint
	Name             string
	Description      string
	StartingLocation LocationResponse
}
