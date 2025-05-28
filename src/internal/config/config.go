package config

import (
	"github.com/charmbracelet/bubbles/key"
)

type Config struct {
	BaseURL  string
	Language string
	KeyMap   KeyMap
}

func NewConfig() *Config {
	c := &Config{}

	c.BaseURL = "http://127.0.0.1:3000"
	c.Language = "en"

	c.KeyMap = KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move right"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit"),
		),
	}

	return c
}
