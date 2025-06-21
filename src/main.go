package main

import (
	"embed"
	"farental/core/request"
	"farental/internal/config"
	"farental/internal/context"
	"farental/internal/keybind"
	"farental/internal/keymapmanager"
	"farental/internal/lang"
	"farental/model"
	"farental/model/activityselection"
	"farental/model/charactercreation"
	"farental/model/characterselection"
	"farental/model/chat"
	"farental/model/craftselection"
	"farental/model/fightselection"
	"farental/model/gamedashboard"
	"farental/model/inventory"
	"farental/model/login"
	"farental/model/travelselection"
	"farental/style"
	"github.com/charmbracelet/lipgloss"
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

	registerContents()

	registerKeymapContexts()

	context.ContentManager.SwitchContent(nil, model.ContentLogin) // ContentLogin

	p := tea.NewProgram(context.ContentManager.GetCurrentModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func registerContents() {
	context.ContentManager.RegisterContent(model.ContentLogin, login.New())
	context.ContentManager.RegisterContent(model.ContentCharacterSelection, characterselection.New())
	context.ContentManager.RegisterContent(model.ContentCharacterCreation, charactercreation.New())
	context.ContentManager.RegisterContent(model.ContentGameDashboard, gamedashboard.New())
	context.ContentManager.RegisterContent(model.ContentActivitySelection, activityselection.New())
	context.ContentManager.RegisterContent(model.ContentTravelSelection, travelselection.New())
	context.ContentManager.RegisterContent(model.ContentFightSelection, fightselection.New())
	context.ContentManager.RegisterContent(model.ContentCraftSelection, craftselection.New())
	context.ContentManager.RegisterContent(model.ContentInventory, inventory.New())
	context.ContentManager.RegisterContent(model.ContentChat, chat.New())
}

func registerKeymapContexts() {
	mainHelpStyle := keymapmanager.Style{
		EssentialKey:               style.NeutralDimTextStyle.Bold(true),
		EssentialKeyDescription:    style.NeutralDimTextStyle,
		EssentialKeySeparator:      style.NeutralDimTextStyle,
		EssentialKeySeparatorValue: " - ",
		EssentialColSeparator:      style.NeutralDimTextStyle,
		EssentialColSeparatorValue: " • ",
		FullKey:                    style.NeutralDimTextStyle.Bold(true),
		FullKeyDescription:         style.NeutralDimTextStyle,
		FullKeySeparator:           style.NeutralDimTextStyle,
		FullKeySeparatorValue:      " - ",
		FullColSeparator:           lipgloss.Style{},
		FullColSeparatorValue:      "   ",
	}

	loginKeymap := keymapmanager.NewKeymap(2)
	loginKeymap.Style = mainHelpStyle
	loginKeymap.NewKeyBinding(keybind.Tab, false)
	loginKeymap.NewKeyBinding(keybind.ShiftTab, false)
	loginKeymap.NewKeyBinding(keybind.Submit, true)
	loginKeymap.NewKeyBinding(keybind.Quit, true)
	loginKeymap.NewKeyBinding(keybind.HelpMore, true)

	context.KeymapManager.RegisterContext(model.ContextLogin, loginKeymap)
}
