package config

import (
	"github.com/charmbracelet/bubbles/key"
)

type ModularKeyMap struct {
	bindings          [][]key.Binding
	essentialBindings []key.Binding
}

func (k ModularKeyMap) ShortHelp() []key.Binding {
	return k.essentialBindings
}

func (k ModularKeyMap) FullHelp() [][]key.Binding {
	return k.bindings
}

func (k *ModularKeyMap) SetBindings(b [][]key.Binding) {
	k.bindings = b
}

func (k *ModularKeyMap) SetEssentialBindings(b []key.Binding) {
	k.essentialBindings = b
}
