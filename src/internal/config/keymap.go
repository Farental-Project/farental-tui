package config

import (
	"github.com/charmbracelet/bubbles/key"
)

type ModularKeyMap struct {
	Bindings          [][]key.Binding
	EssentialBindings []key.Binding
}

func (k ModularKeyMap) ShortHelp() []key.Binding {
	return k.EssentialBindings
}

func (k ModularKeyMap) FullHelp() [][]key.Binding {
	return k.Bindings
}

func (k *ModularKeyMap) SetBindings(b [][]key.Binding) {
	k.Bindings = b
}

func (k *ModularKeyMap) SetEssentialBindings(b []key.Binding) {
	k.EssentialBindings = b
}
