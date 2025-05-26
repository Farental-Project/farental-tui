package model

import (
	tea "github.com/charmbracelet/bubbletea"
)

type InitMsg int

func InitCmd() tea.Msg {
	return InitMsg(0)
}
