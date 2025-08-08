package main

import (
	"embed"
	"farental/core/request"
	"farental/internal/config"
	"farental/internal/context"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/screen"
	"farental/screen/activity"
	"farental/screen/charactercreation"
	"farental/screen/characterselection"
	"farental/screen/chat"
	"farental/screen/craft"
	"farental/screen/dashboard"
	"farental/screen/fight"
	"farental/screen/inventory"
	"farental/screen/login"
	"farental/screen/travel"
	"farental/style"
	"github.com/halsten-dev/bubblehelp"
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

	bubblehelp.Init()

	registerKeymapContexts()

	// Orvyn
	orvyn.Init()

	orvyn.RegisterScreen(screen.IDLogin, login.New())
	orvyn.RegisterScreen(screen.IDCharacterSelection, characterselection.New())
	orvyn.RegisterScreen(screen.IDCharacterCreation, charactercreation.New())
	orvyn.RegisterScreen(screen.IDDashBoard, dashboard.New())
	orvyn.RegisterScreen(screen.IDTravel, travel.New())
	orvyn.RegisterScreen(screen.IDActivity, activity.New())
	orvyn.RegisterScreen(screen.IDFight, fight.New())
	orvyn.RegisterScreen(screen.IDCraft, craft.New())
	orvyn.RegisterScreen(screen.IDChat, chat.New())
	orvyn.RegisterScreen(screen.IDInventory, inventory.New())
	orvyn.SwitchScreen(screen.IDLogin)

	p := tea.NewProgram(&App{}, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func registerKeymapContexts() {
	mainHelpStyle := style.MainHelpStyle

	loginKeymap := bubblehelp.NewKeymap(2)
	loginKeymap.Style = mainHelpStyle
	loginKeymap.NewKeyBinding(keybind.Tab, false)
	loginKeymap.NewKeyBinding(keybind.ShiftTab, false)
	loginKeymap.NewKeyBinding(keybind.Enter, true)
	loginKeymap.NewKeyBinding(keybind.Quit, true)
	loginKeymap.NewKeyBinding(keybind.Help, true)

	bubblehelp.RegisterContext(keybind.ContextLogin, loginKeymap)

	characterSelectionKeymap := bubblehelp.NewKeymap(2)
	characterSelectionKeymap.Style = mainHelpStyle
	characterSelectionKeymap.NewKeyBinding(keybind.Up, false)
	characterSelectionKeymap.NewKeyBinding(keybind.Down, false)
	characterSelectionKeymap.NewKeyBinding(keybind.NKey, true)
	characterSelectionKeymap.NewKeyBinding(keybind.Enter, true)
	characterSelectionKeymap.SetHelpDesc(keybind.NKey, lang.L("new character"))
	characterSelectionKeymap.NewKeyBinding(keybind.Esc, true)
	characterSelectionKeymap.SetHelpDesc(keybind.Esc, lang.L("logout"))
	characterSelectionKeymap.NewKeyBinding(keybind.Quit, false)
	characterSelectionKeymap.NewKeyBinding(keybind.Help, true)

	bubblehelp.RegisterContext(keybind.ContextCharacterSel, characterSelectionKeymap)

	characterCreationKeymap := bubblehelp.NewKeymap(2)
	characterCreationKeymap.Style = mainHelpStyle
	characterCreationKeymap.NewKeyBinding(keybind.Enter, true)
	characterCreationKeymap.NewKeyBinding(keybind.Esc, true)
	characterCreationKeymap.NewKeyBinding(keybind.Quit, true)

	bubblehelp.RegisterContext(keybind.ContextCharacterCreation, characterCreationKeymap)

	gameDashboardKeymap := bubblehelp.NewKeymap(2)
	gameDashboardKeymap.Style = mainHelpStyle
	gameDashboardKeymap.NewKeyBinding(keybind.TKey, false)
	gameDashboardKeymap.SetHelpDesc(keybind.TKey, lang.L("travels"))
	gameDashboardKeymap.NewKeyBinding(keybind.AKey, false)
	gameDashboardKeymap.SetHelpDesc(keybind.AKey, lang.L("activities"))
	gameDashboardKeymap.NewKeyBinding(keybind.CKey, false)
	gameDashboardKeymap.SetHelpDesc(keybind.CKey, lang.L("crafts"))
	gameDashboardKeymap.NewKeyBinding(keybind.FKey, false)
	gameDashboardKeymap.SetHelpDesc(keybind.FKey, lang.L("fights"))
	gameDashboardKeymap.NewKeyBinding(keybind.YKey, false)
	gameDashboardKeymap.SetHelpDesc(keybind.YKey, lang.L("chat"))
	gameDashboardKeymap.NewKeyBinding(keybind.LocationServices, false)
	gameDashboardKeymap.NewKeyBinding(keybind.Npcs, false)
	gameDashboardKeymap.NewKeyBinding(keybind.Scripts, false)
	gameDashboardKeymap.NewKeyBinding(keybind.IKey, false)
	gameDashboardKeymap.SetHelpDesc(keybind.IKey, lang.L("inventory"))
	gameDashboardKeymap.NewKeyBinding(keybind.Space, true)
	gameDashboardKeymap.NewKeyBinding(keybind.Esc, false)
	gameDashboardKeymap.SetHelpDesc(keybind.Esc, lang.L("character selection"))
	gameDashboardKeymap.NewKeyBinding(keybind.Quit, true)
	gameDashboardKeymap.NewKeyBinding(keybind.Help, true)

	bubblehelp.RegisterContext(keybind.ContextGameDashboard, gameDashboardKeymap)

	filterSelListBasicKeymap := bubblehelp.NewKeymap(3)
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

	bubblehelp.RegisterContext(keybind.ContextFilterSelectionListBasic, filterSelListBasicKeymap)

	filterSelListIncDecKeymap := bubblehelp.NewKeymap(3)
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

	bubblehelp.RegisterContext(keybind.ContextFilterSelectionListIncDec, filterSelListIncDecKeymap)

	filterSelListPageKeymap := bubblehelp.NewKeymap(3)
	filterSelListPageKeymap.Style = mainHelpStyle
	filterSelListPageKeymap.NewKeyBinding(keybind.Up, false)
	filterSelListPageKeymap.NewKeyBinding(keybind.Down, false)
	filterSelListPageKeymap.NewKeyBinding(keybind.Right, false)
	filterSelListPageKeymap.SetHelpDesc(keybind.Right, lang.L("next page"))
	filterSelListPageKeymap.NewKeyBinding(keybind.Left, false)
	filterSelListPageKeymap.SetHelpDesc(keybind.Left, lang.L("previous page"))
	filterSelListPageKeymap.NewKeyBinding(keybind.GotoListStart, false)
	filterSelListPageKeymap.NewKeyBinding(keybind.GotoListEnd, false)
	filterSelListPageKeymap.NewKeyBinding(keybind.Filter, true)
	filterSelListPageKeymap.NewKeyBinding(keybind.Enter, true)
	filterSelListPageKeymap.NewKeyBinding(keybind.Esc, true)
	filterSelListPageKeymap.NewKeyBinding(keybind.Quit, true)
	filterSelListPageKeymap.NewKeyBinding(keybind.Help, true)

	bubblehelp.RegisterContext(keybind.ContextFilterSelectionListPage, filterSelListPageKeymap)

	filterSelListBasicKeymapWithNew := bubblehelp.NewKeymap(3)
	filterSelListBasicKeymapWithNew.Style = mainHelpStyle
	filterSelListBasicKeymapWithNew.NewKeyBinding(keybind.Up, false)
	filterSelListBasicKeymapWithNew.NewKeyBinding(keybind.Down, false)
	filterSelListBasicKeymapWithNew.NewKeyBinding(keybind.GotoListStart, false)
	filterSelListBasicKeymapWithNew.NewKeyBinding(keybind.GotoListEnd, false)
	filterSelListBasicKeymapWithNew.NewKeyBinding(keybind.NKey, true)
	filterSelListBasicKeymapWithNew.SetHelpDesc(keybind.NKey, lang.L("new"))
	filterSelListBasicKeymapWithNew.NewKeyBinding(keybind.Filter, true)
	filterSelListBasicKeymapWithNew.NewKeyBinding(keybind.Enter, true)
	filterSelListBasicKeymapWithNew.NewKeyBinding(keybind.Esc, true)
	filterSelListBasicKeymapWithNew.NewKeyBinding(keybind.Quit, true)
	filterSelListBasicKeymapWithNew.NewKeyBinding(keybind.Help, true)

	bubblehelp.RegisterContext(keybind.ContextFilterSelectionListWithNew, filterSelListBasicKeymapWithNew)

	craftKeymap := bubblehelp.NewKeymap(3)
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

	bubblehelp.RegisterContext(keybind.ContextCraft, craftKeymap)

	inventoryKeymap := bubblehelp.NewKeymap(3)
	inventoryKeymap.Style = mainHelpStyle
	inventoryKeymap.NewKeyBinding(keybind.Up, false)
	inventoryKeymap.NewKeyBinding(keybind.Down, false)
	inventoryKeymap.NewKeyBinding(keybind.Filter, true)
	inventoryKeymap.NewKeyBinding(keybind.GotoListStart, false)
	inventoryKeymap.NewKeyBinding(keybind.GotoListEnd, false)
	inventoryKeymap.NewKeyBinding(keybind.Use, true)
	inventoryKeymap.NewKeyBinding(keybind.EKey, true)
	inventoryKeymap.SetHelpDesc(keybind.EKey, lang.L("equip"))
	inventoryKeymap.NewKeyBinding(keybind.Enter, true)
	inventoryKeymap.NewKeyBinding(keybind.Esc, true)
	inventoryKeymap.NewKeyBinding(keybind.Quit, true)
	inventoryKeymap.NewKeyBinding(keybind.Help, true)

	bubblehelp.RegisterContext(keybind.ContextInventory, inventoryKeymap)

	chatKeymap := bubblehelp.NewKeymap(3)
	chatKeymap.Style = mainHelpStyle
	chatKeymap.NewKeyBinding(keybind.Enter, true)
	chatKeymap.SetHelpDesc(keybind.Enter, lang.L("send message"))
	chatKeymap.NewKeyBinding(keybind.YKeyCtrl, true)
	chatKeymap.SetHelpDesc(keybind.YKeyCtrl, lang.L("new line"))
	chatKeymap.NewKeyBinding(keybind.Esc, true)
	chatKeymap.NewKeyBinding(keybind.Quit, true)

	bubblehelp.RegisterContext(keybind.ContextChat, chatKeymap)

	characterSheetKeymap := bubblehelp.NewKeymap(3)
	characterSheetKeymap.Style = mainHelpStyle
	characterSheetKeymap.NewKeyBinding(keybind.PrevPage, false)
	characterSheetKeymap.NewKeyBinding(keybind.NextPage, false)
	characterSheetKeymap.NewKeyBinding(keybind.Esc, true)
	characterSheetKeymap.NewKeyBinding(keybind.Quit, true)

	bubblehelp.RegisterContext(keybind.ContextCharacterSheet, characterSheetKeymap)

	locationServicesKeymap := bubblehelp.NewKeymap(2)
	locationServicesKeymap.Style = mainHelpStyle
	locationServicesKeymap.NewKeyBinding(keybind.RKey, true)
	locationServicesKeymap.SetHelpDesc(keybind.RKey, lang.L("sleep in tavern"))
	locationServicesKeymap.NewKeyBinding(keybind.MKey, true)
	locationServicesKeymap.SetHelpDesc(keybind.MKey, lang.L("mailbox"))
	locationServicesKeymap.NewKeyBinding(keybind.Esc, true)
	locationServicesKeymap.SetHelpDesc(keybind.Esc, lang.L("close"))
	locationServicesKeymap.NewKeyBinding(keybind.Quit, true)

	bubblehelp.RegisterContext(keybind.ContextLocationServices, locationServicesKeymap)

	mailReaderKeymap := bubblehelp.NewKeymap(3)
	mailReaderKeymap.Style = mainHelpStyle
	mailReaderKeymap.NewKeyBinding(keybind.PKey, true)
	mailReaderKeymap.SetHelpDesc(keybind.PKey, lang.L("pay the sender"))
	mailReaderKeymap.NewKeyBinding(keybind.TKey, true)
	mailReaderKeymap.SetHelpDesc(keybind.TKey, lang.L("transfer all attachments"))
	mailReaderKeymap.NewKeyBinding(keybind.RKey, true)
	mailReaderKeymap.SetHelpDesc(keybind.RKey, lang.L("read / unread flag"))
	mailReaderKeymap.NewKeyBinding(keybind.Esc, true)
	mailReaderKeymap.NewKeyBinding(keybind.Quit, false)
	mailReaderKeymap.NewKeyBinding(keybind.Help, true)

	bubblehelp.RegisterContext(keybind.ContextMailReader, mailReaderKeymap)

	MailWidgetNormalModeKeymap := bubblehelp.NewKeymap(2)
	MailWidgetNormalModeKeymap.Style = style.MainHelpStyle
	MailWidgetNormalModeKeymap.NewKeyBinding(keybind.EKey, true)
	MailWidgetNormalModeKeymap.SetHelpDesc(keybind.EKey, lang.L("edit"))
	MailWidgetNormalModeKeymap.NewKeyBinding(keybind.Tab, false)
	MailWidgetNormalModeKeymap.NewKeyBinding(keybind.ShiftTab, false)
	MailWidgetNormalModeKeymap.NewKeyBinding(keybind.Esc, true)
	MailWidgetNormalModeKeymap.NewKeyBinding(keybind.Quit, false)
	MailWidgetNormalModeKeymap.NewKeyBinding(keybind.Help, true)

	bubblehelp.RegisterContext(keybind.ContextMailWidgetNormalMode, MailWidgetNormalModeKeymap)
}
