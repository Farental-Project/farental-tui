package keybind

import (
	"farental/art"
	"farental/internal/style"
	"fmt"

	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
)

const (
	ContextBackAndQuit                    bubblehelp.KeymapContext = "backAndQuit"
	ContextLogin                          bubblehelp.KeymapContext = "login"
	ContextCharacterSel                   bubblehelp.KeymapContext = "characterSelection"
	ContextCharacterCreation              bubblehelp.KeymapContext = "characterCreation"
	ContextCharacterSheet                 bubblehelp.KeymapContext = "characterSheet"
	ContextGameDashboard                  bubblehelp.KeymapContext = "gameDashboard"
	ContextFilterSelectionListBasic       bubblehelp.KeymapContext = "filterSelectionListBasic"
	ContextFilterSelectionListIncDec      bubblehelp.KeymapContext = "filterSelectionListIncDec"
	ContextFightList                      bubblehelp.KeymapContext = "fightList"
	ContextFilterSelectionListWithNew     bubblehelp.KeymapContext = "filterSelectionListWithNew"
	ContextCraft                          bubblehelp.KeymapContext = "craft"
	ContextInventory                      bubblehelp.KeymapContext = "inventory"
	ContextTravel                         bubblehelp.KeymapContext = "travel"
	ContextChat                           bubblehelp.KeymapContext = "chat"
	ContextLocationServices               bubblehelp.KeymapContext = "locationServices"
	ContextMailBox                        bubblehelp.KeymapContext = "mailBox"
	ContextMailReader                     bubblehelp.KeymapContext = "mailReader"
	ContextMailWidgetNormalMode           bubblehelp.KeymapContext = "mailWidgetNormalMode"
	ContextMailWriterEditMode             bubblehelp.KeymapContext = "mailWriterEditMode"
	ContextMailDetailEditorEditMode       bubblehelp.KeymapContext = "mailDetailEditorEditMode"
	ContextMailDetailEditorAttachmentList bubblehelp.KeymapContext = "mailDetailEditorAttachmentList"
	ContextScriptExplorer                 bubblehelp.KeymapContext = "scriptExplorer"
	ContextScriptEditorWidgetNormalMode   bubblehelp.KeymapContext = "scriptEditorWidgetNormalMode"
	ContextScriptEditorRulesList          bubblehelp.KeymapContext = "scriptEditorRulesList"
	ContextScriptEditorRulesListItem      bubblehelp.KeymapContext = "scriptEditorRulesListItem"
	ContextScriptEditorRuleInspector      bubblehelp.KeymapContext = "scriptEditorRuleInspector"
	ContextScriptAbilitySelection         bubblehelp.KeymapContext = "scriptAbilitySelection"
	ContextBasicEditMode                  bubblehelp.KeymapContext = "basicEditMode"
	ContextBank                           bubblehelp.KeymapContext = "bank"
	ContextNpc                            bubblehelp.KeymapContext = "npc"
	ContextFightHistory                   bubblehelp.KeymapContext = "fightHistory"
	ContextShop                           bubblehelp.KeymapContext = "shop"
)

func InitContexts() {
	mainHelpStyle := style.MainHelpStyle

	backAndQuit := bubblehelp.NewKeymap(2)
	backAndQuit.Style = mainHelpStyle
	backAndQuit.NewKeyBinding(Esc, true)
	backAndQuit.NewKeyBinding(Quit, true)

	bubblehelp.RegisterContext(ContextBackAndQuit, backAndQuit)

	loginKeymap := bubblehelp.NewKeymap(2)
	loginKeymap.Style = mainHelpStyle
	loginKeymap.NewKeyBinding(Tab, false)
	loginKeymap.NewKeyBinding(ShiftTab, false)
	loginKeymap.NewKeyBinding(NKeyCtrl, true)
	loginKeymap.SetHelpDesc(NKeyCtrl, lokyn.L("new account"))
	loginKeymap.NewKeyBinding(Enter, true)
	loginKeymap.NewKeyBinding(Quit, true)
	loginKeymap.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextLogin, loginKeymap)

	characterSelectionKeymap := bubblehelp.NewKeymap(2)
	characterSelectionKeymap.Style = mainHelpStyle
	characterSelectionKeymap.NewKeyBinding(Up, false)
	characterSelectionKeymap.NewKeyBinding(Down, false)
	characterSelectionKeymap.NewKeyBinding(NKey, true)
	characterSelectionKeymap.NewKeyBinding(Enter, true)
	characterSelectionKeymap.SetHelpDesc(NKey, lokyn.L("new character"))
	characterSelectionKeymap.NewKeyBinding(UKey, false)
	characterSelectionKeymap.SetHelpDesc(UKey, lokyn.L("user settings"))
	characterSelectionKeymap.NewKeyBinding(Esc, true)
	characterSelectionKeymap.SetHelpDesc(Esc, lokyn.L("logout"))
	characterSelectionKeymap.NewKeyBinding(Quit, false)
	characterSelectionKeymap.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextCharacterSel, characterSelectionKeymap)

	characterCreationKeymap := bubblehelp.NewKeymap(2)
	characterCreationKeymap.Style = mainHelpStyle
	characterCreationKeymap.NewKeyBinding(Enter, true)
	characterCreationKeymap.NewKeyBinding(Esc, true)
	characterCreationKeymap.NewKeyBinding(Quit, true)

	bubblehelp.RegisterContext(ContextCharacterCreation, characterCreationKeymap)

	gameDashboardKeymap := bubblehelp.NewKeymap(2)
	gameDashboardKeymap.Style = mainHelpStyle
	gameDashboardKeymap.NewKeyBinding(HKey, false)
	gameDashboardKeymap.SetHelpDesc(HKey, lokyn.L("character"))
	gameDashboardKeymap.NewKeyBinding(TKey, false)
	gameDashboardKeymap.SetHelpDesc(TKey, lokyn.L("travels"))
	gameDashboardKeymap.NewKeyBinding(AKey, false)
	gameDashboardKeymap.SetHelpDesc(AKey, lokyn.L("activities"))
	gameDashboardKeymap.NewKeyBinding(CKey, false)
	gameDashboardKeymap.SetHelpDesc(CKey, lokyn.L("crafts"))
	gameDashboardKeymap.NewKeyBinding(FKey, false)
	gameDashboardKeymap.SetHelpDesc(FKey, lokyn.L("fights"))
	gameDashboardKeymap.NewKeyBinding(YKey, false)
	gameDashboardKeymap.SetHelpDesc(YKey, lokyn.L("chat"))
	gameDashboardKeymap.NewKeyBinding(LKey, false)
	gameDashboardKeymap.SetHelpDesc(LKey, lokyn.L("location service"))
	gameDashboardKeymap.NewKeyBinding(MKey, false)
	gameDashboardKeymap.SetHelpDesc(MKey, lokyn.L("more location info"))
	gameDashboardKeymap.NewKeyBinding(NKey, false)
	gameDashboardKeymap.SetHelpDesc(NKey, lokyn.L("npcs"))
	gameDashboardKeymap.NewKeyBinding(SKey, false)
	gameDashboardKeymap.SetHelpDesc(SKey, lokyn.L("scripts"))
	gameDashboardKeymap.NewKeyBinding(IKey, false)
	gameDashboardKeymap.SetHelpDesc(IKey, lokyn.L("inventory"))
	gameDashboardKeymap.NewKeyBinding(Space, true)
	gameDashboardKeymap.NewKeyBinding(UKey, false)
	gameDashboardKeymap.SetHelpDesc(UKey, lokyn.L("user settings"))
	gameDashboardKeymap.NewKeyBinding(Esc, false)
	gameDashboardKeymap.SetHelpDesc(Esc, lokyn.L("character selection"))
	gameDashboardKeymap.NewKeyBinding(Quit, true)
	gameDashboardKeymap.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextGameDashboard, gameDashboardKeymap)

	filterSelListBasicKeymap := bubblehelp.NewKeymap(3)
	filterSelListBasicKeymap.Style = mainHelpStyle
	filterSelListBasicKeymap.NewKeyBinding(Up, false)
	filterSelListBasicKeymap.NewKeyBinding(Down, false)
	filterSelListBasicKeymap.NewKeyBinding(GotoListStart, false)
	filterSelListBasicKeymap.NewKeyBinding(GotoListEnd, false)
	filterSelListBasicKeymap.NewKeyBinding(Filter, true)
	filterSelListBasicKeymap.NewKeyBinding(Enter, true)
	filterSelListBasicKeymap.NewKeyBinding(Esc, true)
	filterSelListBasicKeymap.NewKeyBinding(Quit, true)
	filterSelListBasicKeymap.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextFilterSelectionListBasic, filterSelListBasicKeymap)

	filterSelListIncDecKeymap := bubblehelp.NewKeymap(3)
	filterSelListIncDecKeymap.Style = mainHelpStyle
	filterSelListIncDecKeymap.NewKeyBinding(Up, false)
	filterSelListIncDecKeymap.NewKeyBinding(Down, false)
	filterSelListIncDecKeymap.NewKeyBinding(Left, false)
	filterSelListIncDecKeymap.SetHelpDesc(Left, lokyn.L("decrease"))
	filterSelListIncDecKeymap.NewKeyBinding(Right, false)
	filterSelListIncDecKeymap.SetHelpDesc(Right, lokyn.L("increase"))
	filterSelListIncDecKeymap.NewKeyBinding(GotoListStart, false)
	filterSelListIncDecKeymap.NewKeyBinding(GotoListEnd, false)
	filterSelListIncDecKeymap.NewKeyBinding(Filter, true)
	filterSelListIncDecKeymap.NewKeyBinding(Enter, true)
	filterSelListIncDecKeymap.NewKeyBinding(Esc, true)
	filterSelListIncDecKeymap.NewKeyBinding(Quit, true)
	filterSelListIncDecKeymap.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextFilterSelectionListIncDec, filterSelListIncDecKeymap)

	fightListKeymap := bubblehelp.NewKeymap(3)
	fightListKeymap.Style = mainHelpStyle
	fightListKeymap.NewKeyBinding(Up, false)
	fightListKeymap.NewKeyBinding(Down, false)
	fightListKeymap.NewKeyBinding(Right, false)
	fightListKeymap.SetHelpDesc(Right, lokyn.L("next page"))
	fightListKeymap.NewKeyBinding(Left, false)
	fightListKeymap.SetHelpDesc(Left, lokyn.L("previous page"))
	fightListKeymap.NewKeyBinding(GotoListStart, false)
	fightListKeymap.NewKeyBinding(GotoListEnd, false)
	fightListKeymap.NewKeyBinding(Filter, true)
	fightListKeymap.NewKeyBinding(HKey, true)
	fightListKeymap.SetHelpDesc(HKey, lokyn.L("fight history"))
	fightListKeymap.NewKeyBinding(Enter, true)
	fightListKeymap.NewKeyBinding(Esc, true)
	fightListKeymap.NewKeyBinding(Quit, true)
	fightListKeymap.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextFightList, fightListKeymap)

	filterSelListBasicKeymapWithNew := bubblehelp.NewKeymap(3)
	filterSelListBasicKeymapWithNew.Style = mainHelpStyle
	filterSelListBasicKeymapWithNew.NewKeyBinding(Up, false)
	filterSelListBasicKeymapWithNew.NewKeyBinding(Down, false)
	filterSelListBasicKeymapWithNew.NewKeyBinding(GotoListStart, false)
	filterSelListBasicKeymapWithNew.NewKeyBinding(GotoListEnd, false)
	filterSelListBasicKeymapWithNew.NewKeyBinding(NKey, true)
	filterSelListBasicKeymapWithNew.SetHelpDesc(NKey, lokyn.L("new"))
	filterSelListBasicKeymapWithNew.NewKeyBinding(Filter, true)
	filterSelListBasicKeymapWithNew.NewKeyBinding(Enter, true)
	filterSelListBasicKeymapWithNew.NewKeyBinding(Esc, true)
	filterSelListBasicKeymapWithNew.NewKeyBinding(Quit, true)
	filterSelListBasicKeymapWithNew.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextFilterSelectionListWithNew, filterSelListBasicKeymapWithNew)

	craftKeymap := bubblehelp.NewKeymap(3)
	craftKeymap.Style = mainHelpStyle
	craftKeymap.NewKeyBinding(Up, false)
	craftKeymap.NewKeyBinding(Down, false)
	craftKeymap.NewKeyBinding(Left, false)
	craftKeymap.SetHelpDesc(Left, lokyn.L("decrease"))
	craftKeymap.NewKeyBinding(ShiftLeft, false)
	craftKeymap.SetHelpDesc(ShiftLeft, lokyn.L("decrease 10"))
	craftKeymap.NewKeyBinding(Right, false)
	craftKeymap.SetHelpDesc(Right, lokyn.L("increase"))
	craftKeymap.NewKeyBinding(ShiftRight, false)
	craftKeymap.SetHelpDesc(ShiftRight, lokyn.L("increase 10"))
	craftKeymap.NewKeyBinding(PrevPage, false)
	craftKeymap.NewKeyBinding(NextPage, false)
	craftKeymap.NewKeyBinding(GotoListStart, false)
	craftKeymap.NewKeyBinding(GotoListEnd, false)
	craftKeymap.NewKeyBinding(Filter, true)
	craftKeymap.NewKeyBinding(Enter, true)
	craftKeymap.NewKeyBinding(Esc, true)
	craftKeymap.NewKeyBinding(Quit, true)
	craftKeymap.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextCraft, craftKeymap)

	inventoryKeymap := bubblehelp.NewKeymap(3)
	inventoryKeymap.Style = mainHelpStyle
	inventoryKeymap.NewKeyBinding(Up, false)
	inventoryKeymap.NewKeyBinding(Down, false)
	inventoryKeymap.NewKeyBinding(Filter, true)
	inventoryKeymap.NewKeyBinding(GotoListStart, false)
	inventoryKeymap.NewKeyBinding(GotoListEnd, false)
	inventoryKeymap.NewKeyBinding(UKey, true)
	inventoryKeymap.SetHelpDesc(UKey, lokyn.L("use"))
	inventoryKeymap.NewKeyBinding(EKey, true)
	inventoryKeymap.SetHelpDesc(EKey, lokyn.L("equip"))
	inventoryKeymap.NewKeyBinding(Tab, true)
	inventoryKeymap.SetHelpDesc(Tab, lokyn.L("equipped items"))
	inventoryKeymap.NewKeyBinding(Esc, true)
	inventoryKeymap.NewKeyBinding(Quit, true)
	inventoryKeymap.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextInventory, inventoryKeymap)

	travelKeymap := bubblehelp.NewKeymap(3)
	travelKeymap.Style = mainHelpStyle
	travelKeymap.NewKeyBinding(Up, false)
	travelKeymap.NewKeyBinding(Down, false)
	travelKeymap.NewKeyBinding(GotoListStart, false)
	travelKeymap.NewKeyBinding(GotoListEnd, false)
	travelKeymap.NewKeyBinding(Filter, true)
	travelKeymap.NewKeyBinding(Enter, true)
	travelKeymap.NewKeyBinding(Tab, true)
	travelKeymap.SetHelpDesc(Tab, lokyn.L("travel relays"))
	travelKeymap.NewKeyBinding(Esc, true)
	travelKeymap.NewKeyBinding(Quit, true)
	travelKeymap.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextTravel, travelKeymap)

	chatKeymap := bubblehelp.NewKeymap(3)
	chatKeymap.Style = mainHelpStyle
	chatKeymap.NewKeyBinding(Enter, true)
	chatKeymap.SetHelpDesc(Enter, lokyn.L("send message"))
	chatKeymap.NewKeyBinding(YKeyCtrl, true)
	chatKeymap.SetHelpDesc(YKeyCtrl, lokyn.L("new line"))
	chatKeymap.NewKeyBinding(Esc, true)
	chatKeymap.NewKeyBinding(Quit, true)

	bubblehelp.RegisterContext(ContextChat, chatKeymap)

	characterSheetKeymap := bubblehelp.NewKeymap(3)
	characterSheetKeymap.Style = mainHelpStyle
	characterSheetKeymap.NewKeyBinding(PrevPage, false)
	characterSheetKeymap.NewKeyBinding(NextPage, false)
	characterSheetKeymap.NewKeyBinding(Esc, true)
	characterSheetKeymap.NewKeyBinding(Quit, true)

	bubblehelp.RegisterContext(ContextCharacterSheet, characterSheetKeymap)

	locationServicesKeymap := bubblehelp.NewKeymap(2)
	locationServicesKeymap.Style = mainHelpStyle
	locationServicesKeymap.NewKeyBinding(BKey, true)
	locationServicesKeymap.SetHelpDesc(BKey, lokyn.L("bank"))
	locationServicesKeymap.NewKeyBinding(TKey, true)
	locationServicesKeymap.SetHelpDesc(TKey, fmt.Sprintf(
		lokyn.L("sleep in tavern (cost: 10%c)"), art.CharGrynars))
	locationServicesKeymap.NewKeyBinding(RKey, true)
	locationServicesKeymap.SetHelpDesc(RKey, fmt.Sprintf(
		lokyn.L("regen HP and MP (cost: 20%c)"), art.CharGrynars))
	locationServicesKeymap.NewKeyBinding(SKey, true)
	locationServicesKeymap.SetHelpDesc(SKey, lokyn.L("shop"))
	locationServicesKeymap.NewKeyBinding(MKey, true)
	locationServicesKeymap.SetHelpDesc(MKey, lokyn.L("mailbox"))
	locationServicesKeymap.NewKeyBinding(Esc, true)
	locationServicesKeymap.SetHelpDesc(Esc, lokyn.L("close"))
	locationServicesKeymap.NewKeyBinding(Quit, true)

	bubblehelp.RegisterContext(ContextLocationServices, locationServicesKeymap)

	mailBoxKeymap := bubblehelp.NewKeymap(3)
	mailBoxKeymap.Style = mainHelpStyle
	mailBoxKeymap.NewKeyBinding(Up, false)
	mailBoxKeymap.NewKeyBinding(Down, false)
	mailBoxKeymap.NewKeyBinding(GotoListStart, false)
	mailBoxKeymap.NewKeyBinding(GotoListEnd, false)
	mailBoxKeymap.NewKeyBinding(NKey, true)
	mailBoxKeymap.SetHelpDesc(NKey, lokyn.L("new"))
	mailBoxKeymap.NewKeyBinding(Filter, true)
	mailBoxKeymap.NewKeyBinding(Enter, true)
	mailBoxKeymap.SetHelpDesc(Enter, lokyn.L("read mail"))
	mailBoxKeymap.NewKeyBinding(Esc, true)
	mailBoxKeymap.NewKeyBinding(Quit, true)
	mailBoxKeymap.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextMailBox, mailBoxKeymap)

	mailReaderKeymap := bubblehelp.NewKeymap(3)
	mailReaderKeymap.Style = mainHelpStyle
	mailReaderKeymap.NewKeyBinding(PKey, true)
	mailReaderKeymap.SetHelpDesc(PKey, lokyn.L("pay the sender"))
	mailReaderKeymap.NewKeyBinding(TKey, true)
	mailReaderKeymap.SetHelpDesc(TKey, lokyn.L("transfer all attachments"))
	mailReaderKeymap.NewKeyBinding(RKey, true)
	mailReaderKeymap.SetHelpDesc(RKey, lokyn.L("read / unread flag"))
	mailReaderKeymap.NewKeyBinding(Esc, true)
	mailReaderKeymap.NewKeyBinding(Quit, false)
	mailReaderKeymap.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextMailReader, mailReaderKeymap)

	MailWidgetNormalModeKeymap := bubblehelp.NewKeymap(2)
	MailWidgetNormalModeKeymap.Style = style.MainHelpStyle
	MailWidgetNormalModeKeymap.NewKeyBinding(EKey, true)
	MailWidgetNormalModeKeymap.SetHelpDesc(EKey, lokyn.L("edit"))
	MailWidgetNormalModeKeymap.NewKeyBinding(Enter, true)
	MailWidgetNormalModeKeymap.SetHelpDesc(Enter, lokyn.L("send mail"))
	MailWidgetNormalModeKeymap.NewKeyBinding(Tab, true)
	MailWidgetNormalModeKeymap.NewKeyBinding(ShiftTab, true)
	MailWidgetNormalModeKeymap.NewKeyBinding(Esc, true)
	MailWidgetNormalModeKeymap.NewKeyBinding(Quit, false)

	bubblehelp.RegisterContext(ContextMailWidgetNormalMode, MailWidgetNormalModeKeymap)

	scriptExplorerKeymap := bubblehelp.NewKeymap(3)
	scriptExplorerKeymap.Style = mainHelpStyle
	scriptExplorerKeymap.NewKeyBinding(Up, false)
	scriptExplorerKeymap.NewKeyBinding(Down, false)
	scriptExplorerKeymap.NewKeyBinding(NKey, true)
	scriptExplorerKeymap.SetHelpDesc(NKey, lokyn.L("new"))
	scriptExplorerKeymap.NewKeyBinding(EKey, true)
	scriptExplorerKeymap.SetHelpDesc(EKey, lokyn.L("edit"))
	scriptExplorerKeymap.NewKeyBinding(CKey, true)
	scriptExplorerKeymap.SetHelpDesc(CKey, lokyn.L("copy"))
	scriptExplorerKeymap.NewKeyBinding(DKey, true)
	scriptExplorerKeymap.SetHelpDesc(DKey, lokyn.L("delete"))
	scriptExplorerKeymap.NewKeyBinding(GotoListStart, false)
	scriptExplorerKeymap.NewKeyBinding(GotoListEnd, false)
	scriptExplorerKeymap.NewKeyBinding(Filter, true)
	scriptExplorerKeymap.NewKeyBinding(Enter, true)
	scriptExplorerKeymap.SetHelpDesc(Enter, lokyn.L("set active"))
	scriptExplorerKeymap.NewKeyBinding(Tab, false)
	scriptExplorerKeymap.SetHelpDesc(Tab, lokyn.L("toggle public/own scripts"))
	scriptExplorerKeymap.NewKeyBinding(Esc, true)
	scriptExplorerKeymap.NewKeyBinding(Quit, true)
	scriptExplorerKeymap.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextScriptExplorer, scriptExplorerKeymap)

	ScriptEditorWidgetNormalModeKeymap := bubblehelp.NewKeymap(3)
	ScriptEditorWidgetNormalModeKeymap.Style = style.MainHelpStyle
	ScriptEditorWidgetNormalModeKeymap.NewKeyBinding(Num1Key, false)
	ScriptEditorWidgetNormalModeKeymap.SetHelpDesc(Num1Key, lokyn.L("focus script info"))
	ScriptEditorWidgetNormalModeKeymap.NewKeyBinding(Num2Key, false)
	ScriptEditorWidgetNormalModeKeymap.SetHelpDesc(Num2Key, lokyn.L("focus rule list"))
	ScriptEditorWidgetNormalModeKeymap.NewKeyBinding(Num3Key, false)
	ScriptEditorWidgetNormalModeKeymap.SetHelpDesc(Num3Key, lokyn.L("focus rule inspector"))
	ScriptEditorWidgetNormalModeKeymap.NewKeyBinding(EKey, true)
	ScriptEditorWidgetNormalModeKeymap.SetHelpDesc(EKey, lokyn.L("edit"))
	ScriptEditorWidgetNormalModeKeymap.NewKeyBinding(SKeyCtrl, true)
	ScriptEditorWidgetNormalModeKeymap.SetHelpDesc(SKeyCtrl, lokyn.L("save script"))
	ScriptEditorWidgetNormalModeKeymap.NewKeyBinding(Tab, true)
	ScriptEditorWidgetNormalModeKeymap.NewKeyBinding(ShiftTab, true)
	ScriptEditorWidgetNormalModeKeymap.NewKeyBinding(Esc, true)
	ScriptEditorWidgetNormalModeKeymap.NewKeyBinding(Quit, false)
	ScriptEditorWidgetNormalModeKeymap.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextScriptEditorWidgetNormalMode, ScriptEditorWidgetNormalModeKeymap)

	ScriptEditorRulesListKeymap := bubblehelp.NewKeymap(3)
	ScriptEditorRulesListKeymap.Style = style.MainHelpStyle
	ScriptEditorRulesListKeymap.NewKeyBinding(Num1Key, false)
	ScriptEditorRulesListKeymap.SetHelpDesc(Num1Key, lokyn.L("focus script info"))
	ScriptEditorRulesListKeymap.NewKeyBinding(Num2Key, false)
	ScriptEditorRulesListKeymap.SetHelpDesc(Num2Key, lokyn.L("focus rule list"))
	ScriptEditorRulesListKeymap.NewKeyBinding(Num3Key, false)
	ScriptEditorRulesListKeymap.SetHelpDesc(Num3Key, lokyn.L("focus rule inspector"))
	ScriptEditorRulesListKeymap.NewKeyBinding(Up, true)
	ScriptEditorRulesListKeymap.NewKeyBinding(Down, true)
	ScriptEditorRulesListKeymap.NewKeyBinding(ShiftUp, false)
	ScriptEditorRulesListKeymap.SetHelpDesc(ShiftUp, lokyn.L("move rule up"))
	ScriptEditorRulesListKeymap.NewKeyBinding(ShiftDown, false)
	ScriptEditorRulesListKeymap.SetHelpDesc(ShiftDown, lokyn.L("move rule down"))
	ScriptEditorRulesListKeymap.NewKeyBinding(EKey, true)
	ScriptEditorRulesListKeymap.SetHelpDesc(EKey, lokyn.L("edit"))
	ScriptEditorRulesListKeymap.NewKeyBinding(NKey, true)
	ScriptEditorRulesListKeymap.SetHelpDesc(NKey, lokyn.L("new"))
	ScriptEditorRulesListKeymap.NewKeyBinding(IKey, true)
	ScriptEditorRulesListKeymap.SetHelpDesc(IKey, lokyn.L("insert"))
	ScriptEditorRulesListKeymap.NewKeyBinding(CKey, true)
	ScriptEditorRulesListKeymap.SetHelpDesc(CKey, lokyn.L("copy"))
	ScriptEditorRulesListKeymap.NewKeyBinding(DKey, true)
	ScriptEditorRulesListKeymap.SetHelpDesc(DKey, lokyn.L("delete"))
	ScriptEditorRulesListKeymap.NewKeyBinding(SKeyCtrl, true)
	ScriptEditorRulesListKeymap.SetHelpDesc(SKeyCtrl, lokyn.L("save script"))
	ScriptEditorRulesListKeymap.NewKeyBinding(Tab, true)
	ScriptEditorRulesListKeymap.NewKeyBinding(ShiftTab, true)
	ScriptEditorRulesListKeymap.NewKeyBinding(Esc, true)
	ScriptEditorRulesListKeymap.NewKeyBinding(Quit, false)
	ScriptEditorRulesListKeymap.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextScriptEditorRulesList, ScriptEditorRulesListKeymap)

	ScriptEditorRuleInspectorKeymap := bubblehelp.NewKeymap(3)
	ScriptEditorRuleInspectorKeymap.Style = style.MainHelpStyle
	ScriptEditorRuleInspectorKeymap.NewKeyBinding(Tab, true)
	ScriptEditorRuleInspectorKeymap.NewKeyBinding(ShiftTab, true)
	ScriptEditorRuleInspectorKeymap.NewKeyBinding(Esc, true)
	ScriptEditorRuleInspectorKeymap.SetHelpDesc(Esc, lokyn.L("stop editing"))
	ScriptEditorRuleInspectorKeymap.NewKeyBinding(Quit, false)
	ScriptEditorRuleInspectorKeymap.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextScriptEditorRuleInspector, ScriptEditorRuleInspectorKeymap)

	filterScriptAbilitySel := bubblehelp.NewKeymap(3)
	filterScriptAbilitySel.Style = mainHelpStyle
	filterScriptAbilitySel.NewKeyBinding(Tab, true)
	filterScriptAbilitySel.SetHelpDesc(Tab, lokyn.L("description / conditions"))
	filterScriptAbilitySel.NewKeyBinding(Up, false)
	filterScriptAbilitySel.NewKeyBinding(Down, false)
	filterScriptAbilitySel.NewKeyBinding(GotoListStart, false)
	filterScriptAbilitySel.NewKeyBinding(GotoListEnd, false)
	filterScriptAbilitySel.NewKeyBinding(Filter, true)
	filterScriptAbilitySel.NewKeyBinding(Enter, true)
	filterScriptAbilitySel.NewKeyBinding(Esc, true)
	filterScriptAbilitySel.NewKeyBinding(Quit, true)
	filterScriptAbilitySel.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextScriptAbilitySelection, filterScriptAbilitySel)

	BasicEditModeKeymap := bubblehelp.NewKeymap(2)
	BasicEditModeKeymap.Style = style.MainHelpStyle
	BasicEditModeKeymap.NewKeyBinding(Tab, true)
	BasicEditModeKeymap.NewKeyBinding(ShiftTab, true)
	BasicEditModeKeymap.NewKeyBinding(Esc, true)
	BasicEditModeKeymap.SetHelpDesc(Esc, lokyn.L("stop editing"))
	BasicEditModeKeymap.NewKeyBinding(Quit, false)

	bubblehelp.RegisterContext(ContextBasicEditMode, BasicEditModeKeymap)

	BankKeymap := bubblehelp.NewKeymap(2)
	BankKeymap.Style = style.MainHelpStyle
	BankKeymap.NewKeyBinding(TKey, true)
	BankKeymap.SetHelpDesc(TKey, lokyn.L("transfer item"))
	BankKeymap.NewKeyBinding(UKey, true)
	BankKeymap.SetHelpDesc(UKey, lokyn.L("buy upgrade"))
	BankKeymap.NewKeyBinding(Up, false)
	BankKeymap.NewKeyBinding(Down, false)
	BankKeymap.NewKeyBinding(Left, false)
	BankKeymap.SetHelpDesc(Left, lokyn.L("decrease"))
	BankKeymap.NewKeyBinding(ShiftLeft, false)
	BankKeymap.SetHelpDesc(ShiftLeft, lokyn.L("decrease 10"))
	BankKeymap.NewKeyBinding(Right, false)
	BankKeymap.SetHelpDesc(Right, lokyn.L("increase"))
	BankKeymap.NewKeyBinding(ShiftRight, false)
	BankKeymap.SetHelpDesc(ShiftRight, lokyn.L("increase 10"))
	BankKeymap.NewKeyBinding(Esc, true)
	BankKeymap.NewKeyBinding(Quit, false)

	bubblehelp.RegisterContext(ContextBank, BankKeymap)

	NpcKeymap := bubblehelp.NewKeymap(2)
	NpcKeymap.Style = style.MainHelpStyle
	NpcKeymap.NewKeyBinding(Up, false)
	NpcKeymap.NewKeyBinding(Down, false)
	NpcKeymap.NewKeyBinding(Enter, true)
	NpcKeymap.SetHelpDesc(Enter, lokyn.L("speak to npc"))
	NpcKeymap.NewKeyBinding(Esc, true)
	NpcKeymap.NewKeyBinding(Quit, false)

	bubblehelp.RegisterContext(ContextNpc, NpcKeymap)

	FightHistoryKeymap := bubblehelp.NewKeymap(2)
	FightHistoryKeymap.Style = style.MainHelpStyle
	FightHistoryKeymap.NewKeyBinding(Up, false)
	FightHistoryKeymap.NewKeyBinding(Down, false)
	FightHistoryKeymap.NewKeyBinding(Right, false)
	FightHistoryKeymap.SetHelpDesc(Right, lokyn.L("next compo page"))
	FightHistoryKeymap.NewKeyBinding(Left, false)
	FightHistoryKeymap.SetHelpDesc(Left, lokyn.L("previous compo page"))
	FightHistoryKeymap.NewKeyBinding(Enter, true)
	FightHistoryKeymap.SetHelpDesc(Enter, lokyn.L("load log"))
	FightHistoryKeymap.NewKeyBinding(Esc, true)
	FightHistoryKeymap.NewKeyBinding(Quit, false)
	FightHistoryKeymap.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextFightHistory, FightHistoryKeymap)

	ShopKeymap := bubblehelp.NewKeymap(2)
	ShopKeymap.Style = style.MainHelpStyle
	ShopKeymap.NewKeyBinding(Enter, true)
	ShopKeymap.SetHelpDesc(Enter, lokyn.L("buy"))
	ShopKeymap.NewKeyBinding(Up, false)
	ShopKeymap.NewKeyBinding(Down, false)
	ShopKeymap.NewKeyBinding(Tab, true)
	ShopKeymap.SetHelpDesc(Tab, lokyn.L("sell items"))
	ShopKeymap.NewKeyBinding(Esc, true)
	ShopKeymap.NewKeyBinding(Quit, false)

	bubblehelp.RegisterContext(ContextShop, ShopKeymap)
}
