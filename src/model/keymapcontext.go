package model

import (
	"farental/internal/keymapmanager"
)

const (
	ContextLogin                         keymapmanager.KeymapContext = "login"
	ContextCharacterSel                  keymapmanager.KeymapContext = "characterSelection"
	ContextCharacterCreation             keymapmanager.KeymapContext = "characterCreation"
	ContextGameDashboard                 keymapmanager.KeymapContext = "gameDashboard"
	ContextFilterSelectionListBasic      keymapmanager.KeymapContext = "filterSelectionListBasic"
	ContextFilterSelectionListIncDec     keymapmanager.KeymapContext = "filterSelectionListIncDec"
	ContextFilterSelectionListPage       keymapmanager.KeymapContext = "filterSelectionListPage"
	ContextFilterSelectionListIncDecPage keymapmanager.KeymapContext = "filterSelectionListIncDecPage"
)
