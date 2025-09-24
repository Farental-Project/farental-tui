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

type ScriptBody struct {
	ID          uint
	Name        string
	Description string
	IsPrivate   bool
	Rules       []ScriptRuleBody
}

type ScriptRuleBody struct {
	Target       ScriptTarget `validate:"required,valid"`
	Order        int          `validate:"required,min=1"`
	RuleTypeCode string       `validate:"required"`
	AbilityCode  string       `validate:"required"`
	Parameters   string       `validate:"json"`
}

type ScriptBasicResponse struct {
	ID          uint
	Name        string
	Description string
	AuthorName  string
	IsPrivate   bool
}

type ScriptResponse struct {
	ScriptBasicResponse
	Rules []ScriptRuleResponse
}

type ScriptRuleTypeResponse struct {
	Code        string
	Name        string
	Description string
}

type ScriptRuleResponse struct {
	ScriptRuleBody
	ParamStruct ScriptRuleParamStructResponse
}

type ScriptRuleParamStructResponse struct {
	Parameters []ScriptRuleParam
}

type ScriptRuleParam struct {
	Name string
	Type uint
}
