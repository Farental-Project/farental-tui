package main

import (
	"farental/core/request"
	"farental/internal"
	"farental/model"
	"farental/model/characterselection"
	"farental/model/login"
	tea "github.com/charmbracelet/bubbletea"
	"log"
)

func main() {
	appCtx := internal.NewAppCtx()

	request.Init(appCtx)

	f, err := tea.LogToFile("debug.log", "debug")

	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	appCtx.ContentManager.RegisterContent(model.ContentLogin, login.New(appCtx))
	appCtx.ContentManager.RegisterContent(model.ContentCharacterSelection, characterselection.New(appCtx))

	appCtx.ContentManager.SwitchContent(model.ContentLogin)

	p := tea.NewProgram(appCtx.ContentManager.GetCurrentModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
