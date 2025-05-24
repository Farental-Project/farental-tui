package main

import (
	"farental/core/request"
	"farental/data"
	"farental/model"
	tea "github.com/charmbracelet/bubbletea"
	"log"
)

func main() {
	appCtx := data.NewAppCtx()

	request.Init(appCtx)

	f, err := tea.LogToFile("debug.log", "debug")

	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	p := tea.NewProgram(model.AppModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
