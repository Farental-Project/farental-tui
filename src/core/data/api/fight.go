package api

import (
	"time"
)

type FightResponse struct {
	ID uint

	ResolvedTimestamp time.Time

	EventLog EventLogResponse

	Composition FightCompositionResponse
}

type FightCompositionResponse struct {
	ID uint

	Actors []FightActorResponse
}

type FightActorResponse struct {
	Name  string
	Power int
}
