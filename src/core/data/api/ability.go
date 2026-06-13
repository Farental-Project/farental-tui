package api

type AbilityResponse struct {
	ID uint

	Code string

	Name        string
	Description string

	SkillName     string
	SkillLevelMin int
	SkillLevelMax int

	Dice     string
	Cooldown int
	ManaCost int

	TargetGroup      bool
	CanTargetEnemies bool
	CanTargetAllies  bool
	CanTargetSelf    bool

	Conditions []string
}
