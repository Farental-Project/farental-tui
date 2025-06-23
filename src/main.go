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
		EssentialKeySeparatorValue: " ",
		EssentialColSeparator:      style.NeutralDimTextStyle,
		EssentialColSeparatorValue: " • ",
		FullKey:                    style.NeutralLessDimTextStyle.Bold(true),
		FullKeyDescription:         style.NeutralDimTextStyle,
		FullKeySeparator:           style.NeutralDimTextStyle,
		FullKeySeparatorValue:      " ",
		FullColSeparator:           lipgloss.Style{},
		FullColSeparatorValue:      "  ",
	}

	loginKeymap := keymapmanager.NewKeymap(2)
	loginKeymap.Style = mainHelpStyle
	loginKeymap.NewKeyBinding(keybind.Tab, false)
	loginKeymap.NewKeyBinding(keybind.ShiftTab, false)
	loginKeymap.NewKeyBinding(keybind.Enter, true)
	loginKeymap.NewKeyBinding(keybind.Quit, true)
	loginKeymap.NewKeyBinding(keybind.Help, true)

	context.KeymapManager.RegisterContext(model.ContextLogin, loginKeymap)

	characterSelectionKeymap := keymapmanager.NewKeymap(2)
	characterSelectionKeymap.Style = mainHelpStyle
	characterSelectionKeymap.NewKeyBinding(keybind.Up, false)
	characterSelectionKeymap.NewKeyBinding(keybind.Down, false)
	characterSelectionKeymap.NewKeyBinding(keybind.Enter, false)
	characterSelectionKeymap.NewKeyBinding(keybind.NewCharacter, true)
	characterSelectionKeymap.NewKeyBinding(keybind.Esc, true)
	characterSelectionKeymap.NewKeyBinding(keybind.Quit, true)
	characterSelectionKeymap.NewKeyBinding(keybind.Help, true)

	context.KeymapManager.RegisterContext(model.ContextCharacterSel, characterSelectionKeymap)

	characterCreationKeymap := keymapmanager.NewKeymap(2)
	characterCreationKeymap.Style = mainHelpStyle
	characterCreationKeymap.NewKeyBinding(keybind.Tab, true)
	characterCreationKeymap.NewKeyBinding(keybind.ShiftTab, true)
	characterCreationKeymap.NewKeyBinding(keybind.Enter, true)
	characterCreationKeymap.NewKeyBinding(keybind.Esc, true)
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
	gameDashboardKeymap.NewKeyBinding(keybind.Space, true)
	gameDashboardKeymap.NewKeyBinding(keybind.Esc, false)
	gameDashboardKeymap.SetHelpDesc(keybind.Esc, lang.L("character selection"))
	gameDashboardKeymap.NewKeyBinding(keybind.Quit, true)
	gameDashboardKeymap.NewKeyBinding(keybind.Help, true)

	context.KeymapManager.RegisterContext(model.ContextGameDashboard, gameDashboardKeymap)

	filterSelListBasicKeymap := keymapmanager.NewKeymap(3)
	filterSelListBasicKeymap.Style = mainHelpStyle
	filterSelListBasicKeymap.NewKeyBinding(keybind.Up, false)
	filterSelListBasicKeymap.NewKeyBinding(keybind.Down, false)
	filterSelListBasicKeymap.NewKeyBinding(keybind.GotoListStart, false)
	filterSelListBasicKeymap.NewKeyBinding(keybind.GotoListEnd, false)
	filterSelListBasicKeymap.NewKeyBinding(keybind.Filter, true)
	filterSelListBasicKeymap.NewKeyBinding(keybind.Enter, true)
	filterSelListBasicKeymap.NewKeyBinding(keybind.Esc, true)
	filterSelListBasicKeymap.NewKeyBinding(keybind.Quit, true)
	filterSelListBasicKeymap.NewKeyBinding(keybind.Help, true)

	context.KeymapManager.RegisterContext(model.ContextFilterSelectionListBasic, filterSelListBasicKeymap)

	filterSelListIncDecKeymap := keymapmanager.NewKeymap(3)
	filterSelListIncDecKeymap.Style = mainHelpStyle
	filterSelListIncDecKeymap.NewKeyBinding(keybind.Up, false)
	filterSelListIncDecKeymap.NewKeyBinding(keybind.Down, false)
	filterSelListIncDecKeymap.NewKeyBinding(keybind.Left, false)
	filterSelListIncDecKeymap.SetHelpDesc(keybind.Left, lang.L("decrease"))
	filterSelListIncDecKeymap.NewKeyBinding(keybind.Right, false)
	filterSelListIncDecKeymap.SetHelpDesc(keybind.Right, lang.L("increase"))
	filterSelListIncDecKeymap.NewKeyBinding(keybind.GotoListStart, false)
	filterSelListIncDecKeymap.NewKeyBinding(keybind.GotoListEnd, false)
	filterSelListIncDecKeymap.NewKeyBinding(keybind.Filter, true)
	filterSelListIncDecKeymap.NewKeyBinding(keybind.Enter, true)
	filterSelListIncDecKeymap.NewKeyBinding(keybind.Esc, true)
	filterSelListIncDecKeymap.NewKeyBinding(keybind.Quit, true)
	filterSelListIncDecKeymap.NewKeyBinding(keybind.Help, true)

	context.KeymapManager.RegisterContext(model.ContextFilterSelectionListIncDec, filterSelListIncDecKeymap)

	filterSelListPageKeymap := keymapmanager.NewKeymap(3)
	filterSelListPageKeymap.Style = mainHelpStyle
	filterSelListPageKeymap.NewKeyBinding(keybind.Up, false)
	filterSelListPageKeymap.NewKeyBinding(keybind.Down, false)
	filterSelListPageKeymap.NewKeyBinding(keybind.PrevPage, false)
	filterSelListPageKeymap.NewKeyBinding(keybind.NextPage, false)
	filterSelListPageKeymap.NewKeyBinding(keybind.GotoListStart, false)
	filterSelListPageKeymap.NewKeyBinding(keybind.GotoListEnd, false)
	filterSelListPageKeymap.NewKeyBinding(keybind.Filter, true)
	filterSelListPageKeymap.NewKeyBinding(keybind.Enter, true)
	filterSelListPageKeymap.NewKeyBinding(keybind.Esc, true)
	filterSelListPageKeymap.NewKeyBinding(keybind.Quit, true)
	filterSelListPageKeymap.NewKeyBinding(keybind.Help, true)

	context.KeymapManager.RegisterContext(model.ContextFilterSelectionListPage, filterSelListPageKeymap)

	craftKeymap := keymapmanager.NewKeymap(3)
	craftKeymap.Style = mainHelpStyle
	craftKeymap.NewKeyBinding(keybind.Up, false)
	craftKeymap.NewKeyBinding(keybind.Down, false)
	craftKeymap.NewKeyBinding(keybind.Left, false)
	craftKeymap.SetHelpDesc(keybind.Left, lang.L("decrease"))
	craftKeymap.NewKeyBinding(keybind.ShiftLeft, false)
	craftKeymap.SetHelpDesc(keybind.ShiftLeft, lang.L("decrease 10"))
	craftKeymap.NewKeyBinding(keybind.Right, false)
	craftKeymap.SetHelpDesc(keybind.Right, lang.L("increase"))
	craftKeymap.NewKeyBinding(keybind.ShiftRight, false)
	craftKeymap.SetHelpDesc(keybind.ShiftRight, lang.L("increase 10"))
	craftKeymap.NewKeyBinding(keybind.PrevPage, false)
	craftKeymap.NewKeyBinding(keybind.NextPage, false)
	craftKeymap.NewKeyBinding(keybind.GotoListStart, false)
	craftKeymap.NewKeyBinding(keybind.GotoListEnd, false)
	craftKeymap.NewKeyBinding(keybind.Filter, true)
	craftKeymap.NewKeyBinding(keybind.Enter, true)
	craftKeymap.NewKeyBinding(keybind.Esc, true)
	craftKeymap.NewKeyBinding(keybind.Quit, true)
	craftKeymap.NewKeyBinding(keybind.Help, true)

	context.KeymapManager.RegisterContext(model.ContextCraft, craftKeymap)

	inventoryKeymap := keymapmanager.NewKeymap(3)
	inventoryKeymap.Style = mainHelpStyle
	inventoryKeymap.NewKeyBinding(keybind.Up, false)
	inventoryKeymap.NewKeyBinding(keybind.Down, false)
	inventoryKeymap.NewKeyBinding(keybind.Filter, true)
	inventoryKeymap.NewKeyBinding(keybind.GotoListStart, false)
	inventoryKeymap.NewKeyBinding(keybind.GotoListEnd, false)
	inventoryKeymap.NewKeyBinding(keybind.Use, true)
	inventoryKeymap.NewKeyBinding(keybind.Equip, true)
	inventoryKeymap.NewKeyBinding(keybind.Enter, true)
	inventoryKeymap.NewKeyBinding(keybind.Esc, true)
	inventoryKeymap.NewKeyBinding(keybind.Quit, true)
	inventoryKeymap.NewKeyBinding(keybind.Help, true)

	context.KeymapManager.RegisterContext(model.ContextInventory, inventoryKeymap)

	chatKeymap := keymapmanager.NewKeymap(3)
	chatKeymap.Style = mainHelpStyle
	chatKeymap.NewKeyBinding(keybind.Enter, true)
	chatKeymap.SetHelpDesc(keybind.Enter, lang.L("send message"))
	chatKeymap.NewKeyBinding(keybind.NewLine, true)
	chatKeymap.NewKeyBinding(keybind.Esc, true)
	chatKeymap.NewKeyBinding(keybind.Quit, true)
	chatKeymap.NewKeyBinding(keybind.Help, true)

	context.KeymapManager.RegisterContext(model.ContextChat, chatKeymap)
}
