package main

import (
	"embed"
	"farental/core/request"
	"farental/internal"
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
	appCtx := internal.NewAppCtx()

	request.Init(appCtx)

	lang.Init()
	err := lang.AddTranslationFS(translations, "translations")

	if err != nil {
		log.Fatal(err)
	}

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
