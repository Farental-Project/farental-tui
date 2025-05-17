package api

type TaskResponse struct {
	Title              string
	IsRunning          bool
	RemainingTimeHours float64
}
