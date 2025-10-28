package api

type ScriptTarget int

const (
	TargetSelf ScriptTarget = iota
	TargetAllies
	TargetEnemies
)

func (st ScriptTarget) IsValid() bool {
	switch st {
	case TargetSelf, TargetAllies, TargetEnemies:
		return true
	}

	return false
}

type ScriptCountResponse struct {
	Current int
	Max     int
}

type ScriptBody struct {
	ID          uint
	Name        string
	Description string
	IsPrivate   bool
	Rules       []ScriptRuleBody
}

type ScriptRuleBody struct {
	Order          int           `validate:"required,min=1"`
	RuleTypeCode   string        `validate:"required"`
	RuleTypeTarget *ScriptTarget `validate:"valid"`
	AbilityCode    string        `validate:"required"`
	AbilityTarget  ScriptTarget  `validate:"required,valid"`
	Parameters     []ScriptRuleTypeParam
}

type ScriptBasicResponse struct {
	ID          uint
	Name        string
	Description string
	AuthorName  string
	IsPrivate   bool
	IsEditable  bool

	// Only used client side
	IsDuplicated bool
}

type ScriptResponse struct {
	ScriptBasicResponse
	Rules []ScriptRuleBody
}

type ScriptRuleTypeParam struct {
	Identifier string
	Value      string
}

type ScriptRuleTypeStructParam struct {
	Name           string
	Identifier     string
	Type           uint
	PossibleValues []ScriptRuleTypePossibleValue
}

type ScriptRuleTypePossibleValue struct {
	Key   string
	Value string
}

type ScriptRuleTypeResponse struct {
	Code        string
	Name        string
	Description string
}
