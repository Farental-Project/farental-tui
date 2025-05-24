package model

import (
	"farental/data"
	tea "github.com/charmbracelet/bubbletea"
)

type App struct {
	login Login

	ctx *data.AppCtx
}

func AppModel(ctx *data.AppCtx) App {
	return App{
		login: LoginModel(ctx),
		ctx:   ctx,
	}
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	a.login, cmd = a.login.Update(msg)

	return a, cmd
}

func (a App) View() string {
	return a.login.View()
}
