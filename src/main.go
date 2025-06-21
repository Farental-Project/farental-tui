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
		EssentialKey:               style.NeutralLessDimTextStyle.Bold(true),
		EssentialKeyDescription:    style.NeutralDimTextStyle,
		EssentialKeySeparator:      style.NeutralDimTextStyle,
		EssentialKeySeparatorValue: " - ",
		EssentialColSeparator:      style.NeutralDimTextStyle,
		EssentialColSeparatorValue: " • ",
		FullKey:                    style.NeutralLessDimTextStyle.Bold(true),
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
	loginKeymap.NewKeyBinding(keybind.Help, true)

	context.KeymapManager.RegisterContext(model.ContextLogin, loginKeymap)

	characterSelectionKeymap := keymapmanager.NewKeymap(2)
	characterSelectionKeymap.Style = mainHelpStyle
	characterSelectionKeymap.NewKeyBinding(keybind.Up, false)
	characterSelectionKeymap.NewKeyBinding(keybind.Down, false)
	characterSelectionKeymap.NewKeyBinding(keybind.Submit, false)
	characterSelectionKeymap.NewKeyBinding(keybind.NewCharacter, true)
	characterSelectionKeymap.NewKeyBinding(keybind.Back, true)
	characterSelectionKeymap.NewKeyBinding(keybind.Quit, true)
	characterSelectionKeymap.NewKeyBinding(keybind.Help, true)

	context.KeymapManager.RegisterContext(model.ContextCharacterSel, characterSelectionKeymap)

	characterCreationKeymap := keymapmanager.NewKeymap(2)
	characterCreationKeymap.Style = mainHelpStyle
	characterCreationKeymap.NewKeyBinding(keybind.Tab, true)
	characterCreationKeymap.NewKeyBinding(keybind.ShiftTab, true)
	characterCreationKeymap.NewKeyBinding(keybind.Submit, true)
	characterCreationKeymap.NewKeyBinding(keybind.Back, true)
	characterCreationKeymap.NewKeyBinding(keybind.Quit, true)

	context.KeymapManager.RegisterContext(model.ContextCharacterCreation, characterCreationKeymap)

	gameDashboardKeymap := keymapmanager.NewKeymap(2)
	gameDashboardKeymap.Style = mainHelpStyle
	gameDashboardKeymap.NewKeyBinding(keybind.Travels, false)
	gameDashboardKeymap.NewKeyBinding(keybind.Activities, false)
	gameDashboardKeymap.NewKeyBinding(keybind.Crafts, false)
	gameDashboardKeymap.NewKeyBinding(keybind.Fights, false)
	gameDashboardKeymap.NewKeyBinding(keybind.LocationServices, false)
	gameDashboardKeymap.NewKeyBinding(keybind.Npcs, false)
	gameDashboardKeymap.NewKeyBinding(keybind.Scripts, false)
	gameDashboardKeymap.NewKeyBinding(keybind.Inventory, false)
	gameDashboardKeymap.NewKeyBinding(keybind.Claim, true)
	gameDashboardKeymap.NewKeyBinding(keybind.ChangeCharacter, false)
	gameDashboardKeymap.NewKeyBinding(keybind.Quit, true)
	gameDashboardKeymap.NewKeyBinding(keybind.Help, true)

	context.KeymapManager.RegisterContext(model.ContextGameDashboard, gameDashboardKeymap)

}
