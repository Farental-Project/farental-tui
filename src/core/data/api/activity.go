package api

type ActivityStartBody struct {
	ActivityID uint
	DurationID uint
}

type ActivityResponse struct {
	ID          uint
	Name        string
	Description string

	Skill SkillResponse

	Duration DurationTemplateResponse
}

type DurationTemplateResponse struct {
	ID uint

	Name string

	Durations []DurationResponse
}

type DurationResponse struct {
	ID       uint
	Duration float64
}
