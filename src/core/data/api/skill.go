package api

type SkillResponse struct {
	ID              uint
	Code            string
	Name            string
	Description     string
	IsFightingSkill bool
	MaxLevel        int
}
