package api

type NpcResponse struct {
	ID          uint
	FirstName   string
	LastName    string
	Name        string
	RaceName    string
	Description string
}

type NpcDialogResponse struct {
	Dialog string
}
