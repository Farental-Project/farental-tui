package main

import (
	"embed"
	"farental/art"
	"farental/core/data"
	"farental/core/request"
	"farental/internal/config"
	"farental/internal/context"
	"farental/internal/keybind"
	"farental/internal/style"
	ftheme "farental/internal/theme"
	"farental/screen"
	"farental/screen/activity"
	"farental/screen/bank"
	"farental/screen/charactercreation"
	"farental/screen/characterselection"
	"farental/screen/charactersheet"
	"farental/screen/chat"
	"farental/screen/craft"
	"farental/screen/dashboard"
	"farental/screen/fight"
	"farental/screen/inventory"
	"farental/screen/login"
	"farental/screen/mailbox"
	"farental/screen/maileditor"
	"farental/screen/mailreader"
	"farental/screen/npc"
	"farental/screen/scripteditor"
	"farental/screen/scriptexplorer"
	"farental/screen/shop"
	"farental/screen/travel"
	"fmt"
	"log"

	"github.com/halsten-dev/orvyn"

	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/spf13/viper"

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
	lokyn.Init()
	err = lokyn.AddTranslationFS(translations, "translations")

	if err != nil {
		log.Fatal(err)
	}

	lokyn.SetLanguage(viper.GetString("language"))

	keybind.Init()

	bubblehelp.Init()

	data.InitTargets()

	// Orvyn
	orvyn.Init()

	orvyn.SetTheme(ftheme.FarentalTheme{})

	style.InitHelpStyle()

	registerKeymapContexts()

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
	orvyn.RegisterScreen(screen.IDCharacterSheet, charactersheet.New())
	orvyn.RegisterScreen(screen.IDMailBox, mailbox.New())
	orvyn.RegisterScreen(screen.IDMailReader, mailreader.New())
	orvyn.RegisterScreen(screen.IDMailEditor, maileditor.New())
	orvyn.RegisterScreen(screen.IDScriptExplorer, scriptexplorer.New())
	orvyn.RegisterScreen(screen.IDScriptEditor, scripteditor.New())
	orvyn.RegisterScreen(screen.IDBank, bank.New())
	orvyn.RegisterScreen(screen.IDNpc, npc.New())
	orvyn.RegisterScreen(screen.IDShop, shop.New())

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
	characterSelectionKeymap.SetHelpDesc(keybind.NKey, lokyn.L("new character"))
	characterSelectionKeymap.NewKeyBinding(keybind.Esc, true)
	characterSelectionKeymap.SetHelpDesc(keybind.Esc, lokyn.L("logout"))
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
	gameDashboardKeymap.NewKeyBinding(keybind.HKey, false)
	gameDashboardKeymap.SetHelpDesc(keybind.HKey, lokyn.L("character"))
	gameDashboardKeymap.NewKeyBinding(keybind.TKey, false)
	gameDashboardKeymap.SetHelpDesc(keybind.TKey, lokyn.L("travels"))
	gameDashboardKeymap.NewKeyBinding(keybind.AKey, false)
	gameDashboardKeymap.SetHelpDesc(keybind.AKey, lokyn.L("activities"))
	gameDashboardKeymap.NewKeyBinding(keybind.CKey, false)
	gameDashboardKeymap.SetHelpDesc(keybind.CKey, lokyn.L("crafts"))
	gameDashboardKeymap.NewKeyBinding(keybind.FKey, false)
	gameDashboardKeymap.SetHelpDesc(keybind.FKey, lokyn.L("fights"))
	gameDashboardKeymap.NewKeyBinding(keybind.YKey, false)
	gameDashboardKeymap.SetHelpDesc(keybind.YKey, lokyn.L("chat"))
	gameDashboardKeymap.NewKeyBinding(keybind.LKey, false)
	gameDashboardKeymap.SetHelpDesc(keybind.LKey, lokyn.L("location service"))
	gameDashboardKeymap.NewKeyBinding(keybind.NKey, false)
	gameDashboardKeymap.SetHelpDesc(keybind.NKey, lokyn.L("npcs"))
	gameDashboardKeymap.NewKeyBinding(keybind.SKey, false)
	gameDashboardKeymap.SetHelpDesc(keybind.SKey, lokyn.L("scripts"))
	gameDashboardKeymap.NewKeyBinding(keybind.IKey, false)
	gameDashboardKeymap.SetHelpDesc(keybind.IKey, lokyn.L("inventory"))
	gameDashboardKeymap.NewKeyBinding(keybind.Space, true)
	gameDashboardKeymap.NewKeyBinding(keybind.Esc, false)
	gameDashboardKeymap.SetHelpDesc(keybind.Esc, lokyn.L("character selection"))
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
	filterSelListIncDecKeymap.SetHelpDesc(keybind.Left, lokyn.L("decrease"))
	filterSelListIncDecKeymap.NewKeyBinding(keybind.Right, false)
	filterSelListIncDecKeymap.SetHelpDesc(keybind.Right, lokyn.L("increase"))
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
	filterSelListPageKeymap.SetHelpDesc(keybind.Right, lokyn.L("next page"))
	filterSelListPageKeymap.NewKeyBinding(keybind.Left, false)
	filterSelListPageKeymap.SetHelpDesc(keybind.Left, lokyn.L("previous page"))
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
	filterSelListBasicKeymapWithNew.SetHelpDesc(keybind.NKey, lokyn.L("new"))
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
	craftKeymap.SetHelpDesc(keybind.Left, lokyn.L("decrease"))
	craftKeymap.NewKeyBinding(keybind.ShiftLeft, false)
	craftKeymap.SetHelpDesc(keybind.ShiftLeft, lokyn.L("decrease 10"))
	craftKeymap.NewKeyBinding(keybind.Right, false)
	craftKeymap.SetHelpDesc(keybind.Right, lokyn.L("increase"))
	craftKeymap.NewKeyBinding(keybind.ShiftRight, false)
	craftKeymap.SetHelpDesc(keybind.ShiftRight, lokyn.L("increase 10"))
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
	inventoryKeymap.NewKeyBinding(keybind.UKey, true)
	inventoryKeymap.SetHelpDesc(keybind.UKey, lokyn.L("use"))
	inventoryKeymap.NewKeyBinding(keybind.EKey, true)
	inventoryKeymap.SetHelpDesc(keybind.EKey, lokyn.L("equip"))
	inventoryKeymap.NewKeyBinding(keybind.Tab, true)
	inventoryKeymap.SetHelpDesc(keybind.Tab, lokyn.L("equipped items"))
	inventoryKeymap.NewKeyBinding(keybind.Esc, true)
	inventoryKeymap.NewKeyBinding(keybind.Quit, true)
	inventoryKeymap.NewKeyBinding(keybind.Help, true)

	bubblehelp.RegisterContext(keybind.ContextInventory, inventoryKeymap)

	chatKeymap := bubblehelp.NewKeymap(3)
	chatKeymap.Style = mainHelpStyle
	chatKeymap.NewKeyBinding(keybind.Enter, true)
	chatKeymap.SetHelpDesc(keybind.Enter, lokyn.L("send message"))
	chatKeymap.NewKeyBinding(keybind.YKeyCtrl, true)
	chatKeymap.SetHelpDesc(keybind.YKeyCtrl, lokyn.L("new line"))
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
	locationServicesKeymap.NewKeyBinding(keybind.BKey, true)
	locationServicesKeymap.SetHelpDesc(keybind.BKey, lokyn.L("bank"))
	locationServicesKeymap.NewKeyBinding(keybind.RKey, true)
	locationServicesKeymap.SetHelpDesc(keybind.RKey, fmt.Sprintf(
		lokyn.L("sleep in tavern (cost: 10%c)"), art.CharGrynars))
	locationServicesKeymap.NewKeyBinding(keybind.SKey, true)
	locationServicesKeymap.SetHelpDesc(keybind.SKey, lokyn.L("shop"))
	locationServicesKeymap.NewKeyBinding(keybind.MKey, true)
	locationServicesKeymap.SetHelpDesc(keybind.MKey, lokyn.L("mailbox"))
	locationServicesKeymap.NewKeyBinding(keybind.Esc, true)
	locationServicesKeymap.SetHelpDesc(keybind.Esc, lokyn.L("close"))
	locationServicesKeymap.NewKeyBinding(keybind.Quit, true)

	bubblehelp.RegisterContext(keybind.ContextLocationServices, locationServicesKeymap)

	mailReaderKeymap := bubblehelp.NewKeymap(3)
	mailReaderKeymap.Style = mainHelpStyle
	mailReaderKeymap.NewKeyBinding(keybind.PKey, true)
	mailReaderKeymap.SetHelpDesc(keybind.PKey, lokyn.L("pay the sender"))
	mailReaderKeymap.NewKeyBinding(keybind.TKey, true)
	mailReaderKeymap.SetHelpDesc(keybind.TKey, lokyn.L("transfer all attachments"))
	mailReaderKeymap.NewKeyBinding(keybind.RKey, true)
	mailReaderKeymap.SetHelpDesc(keybind.RKey, lokyn.L("read / unread flag"))
	mailReaderKeymap.NewKeyBinding(keybind.Esc, true)
	mailReaderKeymap.NewKeyBinding(keybind.Quit, false)
	mailReaderKeymap.NewKeyBinding(keybind.Help, true)

	bubblehelp.RegisterContext(keybind.ContextMailReader, mailReaderKeymap)

	MailWidgetNormalModeKeymap := bubblehelp.NewKeymap(2)
	MailWidgetNormalModeKeymap.Style = style.MainHelpStyle
	MailWidgetNormalModeKeymap.NewKeyBinding(keybind.EKey, true)
	MailWidgetNormalModeKeymap.SetHelpDesc(keybind.EKey, lokyn.L("edit"))
	MailWidgetNormalModeKeymap.NewKeyBinding(keybind.Enter, true)
	MailWidgetNormalModeKeymap.SetHelpDesc(keybind.Enter, lokyn.L("send mail"))
	MailWidgetNormalModeKeymap.NewKeyBinding(keybind.Tab, true)
	MailWidgetNormalModeKeymap.NewKeyBinding(keybind.ShiftTab, true)
	MailWidgetNormalModeKeymap.NewKeyBinding(keybind.Esc, true)
	MailWidgetNormalModeKeymap.NewKeyBinding(keybind.Quit, false)

	bubblehelp.RegisterContext(keybind.ContextMailWidgetNormalMode, MailWidgetNormalModeKeymap)

	ScriptEditorWidgetNormalModeKeymap := bubblehelp.NewKeymap(2)
	ScriptEditorWidgetNormalModeKeymap.Style = style.MainHelpStyle
	ScriptEditorWidgetNormalModeKeymap.NewKeyBinding(keybind.EKey, true)
	ScriptEditorWidgetNormalModeKeymap.SetHelpDesc(keybind.EKey, lokyn.L("edit"))
	ScriptEditorWidgetNormalModeKeymap.NewKeyBinding(keybind.SKeyCtrl, true)
	ScriptEditorWidgetNormalModeKeymap.SetHelpDesc(keybind.SKeyCtrl, lokyn.L("save script"))
	ScriptEditorWidgetNormalModeKeymap.NewKeyBinding(keybind.Tab, true)
	ScriptEditorWidgetNormalModeKeymap.NewKeyBinding(keybind.ShiftTab, true)
	ScriptEditorWidgetNormalModeKeymap.NewKeyBinding(keybind.Esc, true)
	ScriptEditorWidgetNormalModeKeymap.NewKeyBinding(keybind.Quit, false)

	bubblehelp.RegisterContext(keybind.ContextScriptEditorWidgetNormalMode, ScriptEditorWidgetNormalModeKeymap)

	ScriptEditorRulesListKeymap := bubblehelp.NewKeymap(2)
	ScriptEditorRulesListKeymap.Style = style.MainHelpStyle
	ScriptEditorRulesListKeymap.NewKeyBinding(keybind.EKey, true)
	ScriptEditorRulesListKeymap.SetHelpDesc(keybind.EKey, lokyn.L("edit"))
	ScriptEditorRulesListKeymap.NewKeyBinding(keybind.NKey, true)
	ScriptEditorRulesListKeymap.SetHelpDesc(keybind.NKey, lokyn.L("new"))
	ScriptEditorRulesListKeymap.NewKeyBinding(keybind.IKey, true)
	ScriptEditorRulesListKeymap.SetHelpDesc(keybind.IKey, lokyn.L("insert"))
	ScriptEditorRulesListKeymap.NewKeyBinding(keybind.DKey, true)
	ScriptEditorRulesListKeymap.SetHelpDesc(keybind.DKey, lokyn.L("delete"))
	ScriptEditorRulesListKeymap.NewKeyBinding(keybind.SKeyCtrl, true)
	ScriptEditorRulesListKeymap.SetHelpDesc(keybind.SKeyCtrl, lokyn.L("save script"))
	ScriptEditorRulesListKeymap.NewKeyBinding(keybind.Tab, true)
	ScriptEditorRulesListKeymap.NewKeyBinding(keybind.ShiftTab, true)
	ScriptEditorRulesListKeymap.NewKeyBinding(keybind.Esc, true)
	ScriptEditorRulesListKeymap.NewKeyBinding(keybind.Quit, false)

	bubblehelp.RegisterContext(keybind.ContextScriptEditorRulesList, ScriptEditorRulesListKeymap)

	BasicEditModeKeymap := bubblehelp.NewKeymap(2)
	BasicEditModeKeymap.Style = style.MainHelpStyle
	BasicEditModeKeymap.NewKeyBinding(keybind.Tab, true)
	BasicEditModeKeymap.NewKeyBinding(keybind.ShiftTab, true)
	BasicEditModeKeymap.NewKeyBinding(keybind.Esc, true)
	BasicEditModeKeymap.SetHelpDesc(keybind.Esc, lokyn.L("stop editing"))
	BasicEditModeKeymap.NewKeyBinding(keybind.Quit, false)

	bubblehelp.RegisterContext(keybind.ContextBasicEditMode, BasicEditModeKeymap)

	BankKeymap := bubblehelp.NewKeymap(2)
	BankKeymap.Style = style.MainHelpStyle
	BankKeymap.NewKeyBinding(keybind.TKey, true)
	BankKeymap.SetHelpDesc(keybind.TKey, lokyn.L("transfer item"))
	BankKeymap.NewKeyBinding(keybind.UKey, true)
	BankKeymap.SetHelpDesc(keybind.UKey, lokyn.L("buy upgrade"))
	BankKeymap.NewKeyBinding(keybind.Up, false)
	BankKeymap.NewKeyBinding(keybind.Down, false)
	BankKeymap.NewKeyBinding(keybind.Left, false)
	BankKeymap.SetHelpDesc(keybind.Left, lokyn.L("decrease"))
	BankKeymap.NewKeyBinding(keybind.ShiftLeft, false)
	BankKeymap.SetHelpDesc(keybind.ShiftLeft, lokyn.L("decrease 10"))
	BankKeymap.NewKeyBinding(keybind.Right, false)
	BankKeymap.SetHelpDesc(keybind.Right, lokyn.L("increase"))
	BankKeymap.NewKeyBinding(keybind.ShiftRight, false)
	BankKeymap.SetHelpDesc(keybind.ShiftRight, lokyn.L("increase 10"))
	BankKeymap.NewKeyBinding(keybind.Esc, true)
	BankKeymap.NewKeyBinding(keybind.Quit, false)

	bubblehelp.RegisterContext(keybind.ContextBank, BankKeymap)

	NpcKeymap := bubblehelp.NewKeymap(2)
	NpcKeymap.Style = style.MainHelpStyle
	NpcKeymap.NewKeyBinding(keybind.Up, false)
	NpcKeymap.NewKeyBinding(keybind.Down, false)
	NpcKeymap.NewKeyBinding(keybind.Enter, true)
	NpcKeymap.SetHelpDesc(keybind.Enter, lokyn.L("speak to npc"))
	NpcKeymap.NewKeyBinding(keybind.Esc, true)
	NpcKeymap.NewKeyBinding(keybind.Quit, false)

	bubblehelp.RegisterContext(keybind.ContextNpc, NpcKeymap)

	ShopKeymap := bubblehelp.NewKeymap(2)
	ShopKeymap.Style = style.MainHelpStyle
	ShopKeymap.NewKeyBinding(keybind.Enter, true)
	ShopKeymap.NewKeyBinding(keybind.Up, false)
	ShopKeymap.NewKeyBinding(keybind.Down, false)
	ShopKeymap.SetHelpDesc(keybind.Enter, lokyn.L("sell items"))
	ShopKeymap.NewKeyBinding(keybind.Esc, true)
	ShopKeymap.NewKeyBinding(keybind.Quit, false)

	bubblehelp.RegisterContext(keybind.ContextShop, ShopKeymap)
}
