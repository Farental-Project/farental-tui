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
	Back = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"))
	Tab = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "Next focus"))
	ShiftTab = key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "Previous focus"))
)
