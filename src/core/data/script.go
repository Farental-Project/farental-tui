package data

import (
	"farental/core/data/api"
	"github.com/halsten-dev/lokyn"
)

type Target struct {
	api.ScriptTarget
}

func (t Target) RenderValue() string {
	switch t.ScriptTarget {
	case api.TargetSelf:
		return lokyn.L("Self")

	case api.TargetAllies:
		return lokyn.L("Allies")

	case api.TargetEnemies:
		return lokyn.L("Enemies")

	}

	return ""
}

var TargetKeys = []string{
	lokyn.L("Self"),
	lokyn.L("Allies"),
	lokyn.L("Enemies"),
}

var Targets = map[string]Target{
	TargetKeys[0]: {
		api.TargetSelf,
	},
	TargetKeys[1]: {
		api.TargetAllies,
	},
	TargetKeys[2]: {
		api.TargetEnemies,
	},
}
