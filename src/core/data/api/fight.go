package api

import (
	"time"
)

type FightStartBody struct {
	ID     uint
	Amount int
}

type FightResponse struct {
	ID                uint
	ResolvedTimestamp time.Time
	Composition       FightCompositionResponse
	Amount            int
}

type FightCompositionResponse struct {
	ID       uint
	Duration Duration
	Simple   bool
	Actors   []FightActorResponse
}

type FightActorResponse struct {
	Name  string
	Power int
}
