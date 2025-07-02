package model

import (
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

type InitMsg int

func InitCmd() tea.Msg {
	return InitMsg(0)
}

type SwitchContentMsg string

func SwitchContentCmd(nextContent string) tea.Cmd {
	return func() tea.Msg {
		return SwitchContentMsg(nextContent)
	}
}

type TickMsg struct {
	Time time.Time
	Tag  uint
}

func TickCmd(m tea.Model, seconds time.Duration, tag uint) tea.Cmd {
	return tea.Tick(seconds*time.Second, func(t time.Time) tea.Msg {
		return TickMsg{
			Time: t,
			Tag:  tag,
		}
	})
}
