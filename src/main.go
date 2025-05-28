package main

import (
	"embed"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/lang"
	"farental/model"
	"farental/model/characterselection"
	"farental/model/login"
	tea "github.com/charmbracelet/bubbletea"
	"log"
)

//go:embed translations
var translations embed.FS

func main() {
	f, err := tea.LogToFile("debug.log", "debug")

	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	context.Init()
	request.Init(context.Client)
	lang.Init()
	err = lang.AddTranslationFS(translations, "translations")

	if err != nil {
		log.Fatal(err)
	}

	context.ContentManager.RegisterContent(model.ContentLogin, login.New())
	context.ContentManager.RegisterContent(model.ContentCharacterSelection, characterselection.New())

	context.ContentManager.SwitchContent(model.ContentLogin)

	p := tea.NewProgram(context.ContentManager.GetCurrentModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
