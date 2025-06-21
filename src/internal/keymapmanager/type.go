package keymapmanager

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type Style struct {
	EssentialKey               lipgloss.Style
	EssentialKeyDescription    lipgloss.Style
	EssentialKeySeparator      lipgloss.Style
	EssentialKeySeparatorValue string
	EssentialColSeparator      lipgloss.Style
	EssentialColSeparatorValue string
	FullKey                    lipgloss.Style
	FullKeyDescription         lipgloss.Style
	FullKeySeparator           lipgloss.Style
	FullKeySeparatorValue      string
	FullColSeparator           lipgloss.Style
	FullColSeparatorValue      string
}

type Key struct {
	Binding   key.Binding
	Essential bool
}

type IKeymap interface {
	EssentialBindings() []Key
	AllBindings() []Key
}

type Keymap struct {
	Bindings           []Key
	ShowAllColumnCount int
	Style              Style
}

func NewKeymap(showAllColCount int) *Keymap {
	return &Keymap{
		Bindings:           make([]Key, 0),
		ShowAllColumnCount: showAllColCount,
		Style: Style{
			EssentialKey: lipgloss.NewStyle().
				Bold(true),
			EssentialKeyDescription: lipgloss.NewStyle().
				Italic(true),
			EssentialKeySeparator: lipgloss.NewStyle().
				Italic(true),
			EssentialKeySeparatorValue: " - ",
			EssentialColSeparator: lipgloss.NewStyle().
				Bold(true),
			EssentialColSeparatorValue: " • ",
			FullKey: lipgloss.NewStyle().
				Bold(true),
			FullKeyDescription: lipgloss.NewStyle().
				Italic(true),
			FullKeySeparator: lipgloss.NewStyle().
				Italic(true),
			FullKeySeparatorValue: " - ",
			FullColSeparator: lipgloss.NewStyle().
				Italic(true),
			FullColSeparatorValue: "   ",
		},
	}
}

func (k *Keymap) NewKeyBinding(binding key.Binding, essential bool) {
	k.Bindings = append(k.Bindings, Key{
		Binding:   binding,
		Essential: essential,
	})
}

func (k *Keymap) EssentialBindings() []Key {
	var essentials []Key

	for _, k := range k.Bindings {
		if !k.Essential {
			continue
		}

		essentials = append(essentials, k)
	}

	return essentials
}

func (k *Keymap) AllBindings() []Key {
	return k.Bindings
}

type KeymapContext string
