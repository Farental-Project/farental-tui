package api

import (
	"time"
)

type EventLogResponse struct {
	Entries []EventLogEntryResponse
}

type EventLogEntryResponse struct {
	Timestamp time.Time
	Order     int
	Value     string
}
