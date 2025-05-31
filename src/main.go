package main

import (
	"embed"
	"farental/core/request"
	"farental/internal/config"
	"farental/internal/context"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/model"
	"farental/model/characterselection"
	"farental/model/gamedashboard"
	"farental/model/login"
	"farental/model/travelselection"
	"github.com/spf13/viper"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

//go:embed translations
var translations embed.FS

func main() {
	f, err := tea.LogToFile("debug.log", "debug")

	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	config.Init()
	context.Init()
	request.Init(context.Client)
	lang.Init()
	err = lang.AddTranslationFS(translations, "translations")

	if err != nil {
		log.Fatal(err)
	}

	lang.SetLanguage(viper.GetString("language"))

	keybind.Init()

	context.ContentManager.RegisterContent(model.ContentLogin, login.New())
	context.ContentManager.RegisterContent(model.ContentCharacterSelection, characterselection.New())
	context.ContentManager.RegisterContent(model.ContentGameDashboard, gamedashboard.New())
	context.ContentManager.RegisterContent(model.ContentTravelSelection, travelselection.New())

	context.ContentManager.SwitchContent(model.ContentLogin) // ContentLogin

	p := tea.NewProgram(context.ContentManager.GetCurrentModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
