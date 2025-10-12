package api

type NpcResponse struct {
	ID        uint
	FirstName string
	LastName  string
	RaceName  string
}

type NpcDialogResponse struct {
	Dialog string
}
