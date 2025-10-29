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

var TargetKeys []string
var Targets map[string]Target

func InitTargets() {
	TargetKeys = []string{
		lokyn.L("Self"),
		lokyn.L("Allies"),
		lokyn.L("Enemies"),
	}

	Targets = map[string]Target{
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
}

func GetTarget(target api.ScriptTarget) Target {
	return Targets[TargetKeys[target]]
}

func GetFilteredTargets(self, allies, enemies bool) ([]string, map[string]Target) {
	var keys []string
	var data map[string]Target
	var key string

	data = make(map[string]Target)

	if self {
		key = TargetKeys[0]
		keys = append(keys, key)
		data[key] = Targets[key]
	}

	if allies {
		key = TargetKeys[1]
		keys = append(keys, key)
		data[key] = Targets[key]
	}

	if enemies {
		key = TargetKeys[2]
		keys = append(keys, key)
		data[key] = Targets[key]
	}

	return keys, data
}
