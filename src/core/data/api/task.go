package api

type TaskResponse struct {
	Title              string
	IsRunning          bool
	RemainingTimeHours float64
}

type Duration struct {
	ID uint

	Duration float64
}
