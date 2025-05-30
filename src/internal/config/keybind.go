package config

import (
	"github.com/charmbracelet/bubbles/key"
)

var (
	Up = key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	)
	Down = key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	)
	Left = key.NewBinding(
		key.WithKeys("left", "l"),
		key.WithHelp("←/h", "move left"),
	)
	Right = key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "move right"),
	)
	Help = key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	)
	Quit = key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	)
	Submit = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	)
	Claim = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "claim"))
	Back = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"))
	Tab = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next focus"))
	ShiftTab = key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "previous focus"))
	Travels = key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "travels"))
	Activities = key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "activities"))
	Crafts = key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "crafts"))
	Fights = key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "fights"))
	Npcs = key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "npcs"))
	Scripts = key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "scripts"))
	LocationServices = key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "location services"))
	Inventory = key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "inventory"))
)
