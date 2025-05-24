package model

import (
	tea "github.com/charmbracelet/bubbletea"
)

type App struct {
	login Login
}

func AppModel() App {
	return App{
		login: LoginModel(),
	}
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return a, nil
}

func (a App) View() string {
	// TODO implement me
	panic("implement me")
}
