package api

type AbilityResponse struct {
	ID uint

	Code string

	Name        string
	Description string

	SkillName string

	Power    int
	Cooldown int
	ManaCost int

	TargetGroup      bool
	CanTargetEnemies bool
	CanTargetAllies  bool
	CanTargetSelf    bool
}
