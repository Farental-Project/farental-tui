package model

import (
	"farental/internal/keymapmanager"
)

const (
	ContextLogin                               keymapmanager.KeymapContext = "login"
	ContextCharacterSel                        keymapmanager.KeymapContext = "characterSelection"
	ContextCharacterCreation                   keymapmanager.KeymapContext = "characterCreation"
	ContextGameDashboard                       keymapmanager.KeymapContext = "gameDashboard"
	ContextFilterSelectionListUnfiltered       keymapmanager.KeymapContext = "filterSelectionListUnfiltered"
	ContextFilterSelectionListFiltered         keymapmanager.KeymapContext = "filterSelectionListFiltered"
	ContextFilterSelectionListFiltering        keymapmanager.KeymapContext = "filterSelectionListFiltering"
	ContextFilterSelectionListIncreaseDecrease keymapmanager.KeymapContext = "filterSelectionListIncreaseDecrease"
)
