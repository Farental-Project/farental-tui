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
	Binding        key.Binding
	Essential      bool
	CustomHelpDesc string
	Visible        bool
}

func (k *Key) GetHelpDesc() string {
	if k.CustomHelpDesc == "" {
		return k.Binding.Help().Desc
	}

	return k.CustomHelpDesc
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
		Binding:        binding,
		Essential:      essential,
		CustomHelpDesc: "",
		Visible:        true,
	})
}

func (k *Keymap) EssentialBindings() []Key {
	var essentials []Key

	for _, k := range k.Bindings {
		if !k.Essential || !k.Visible {
			continue
		}

		essentials = append(essentials, k)
	}

	return essentials
}

func (k *Keymap) AllBindings() []Key {
	var all []Key

	for _, k := range k.Bindings {
		if !k.Visible {
			continue
		}

		all = append(all, k)
	}

	return all
}

func (k *Keymap) Reset() {
	for i := 0; i < len(k.Bindings); i++ {
		k.Bindings[i].CustomHelpDesc = ""
		k.Bindings[i].Visible = true
	}
}

func (k *Keymap) UpdateHelpDesc(keybind key.Binding, desc string) {
	for i := 0; i < len(k.Bindings); i++ {
		if keybind.Help().Key == k.Bindings[i].Binding.Help().Key {
			k.Bindings[i].CustomHelpDesc = desc
			return
		}
	}
}

func (k *Keymap) SetHelpDesc(keybind key.Binding, desc string) {
	for i := 0; i < len(k.Bindings); i++ {
		if keybind.Help().Key == k.Bindings[i].Binding.Help().Key {
			k.Bindings[i].Binding.SetHelp(k.Bindings[i].Binding.Help().Key, desc)
			return
		}
	}
}

func (k *Keymap) SetVisible(keybind key.Binding, visible bool) {
	for i := 0; i < len(k.Bindings); i++ {
		if keybind.Help().Key == k.Bindings[i].Binding.Help().Key {
			k.Bindings[i].Visible = visible
			return
		}
	}
}

type KeymapContext string
