package keybind

import (
	"farental/internal/lang"
	"github.com/charmbracelet/bubbles/key"
)

var (
	Up            key.Binding
	Down          key.Binding
	Left          key.Binding
	Right         key.Binding
	ShiftLeft     key.Binding
	ShiftRight    key.Binding
	Help          key.Binding
	Quit          key.Binding
	Enter         key.Binding
	Space         key.Binding
	Esc           key.Binding
	Filter        key.Binding
	Tab           key.Binding
	ShiftTab      key.Binding
	Npcs          key.Binding
	Scripts       key.Binding
	HKey          key.Binding
	LKey          key.Binding
	GotoListStart key.Binding
	GotoListEnd   key.Binding
	PrevPage      key.Binding
	NextPage      key.Binding
	AKey          key.Binding
	CKey          key.Binding
	DKey          key.Binding
	EKey          key.Binding
	FKey          key.Binding
	IKey          key.Binding
	MKey          key.Binding
	NKey          key.Binding
	PKey          key.Binding
	RKey          key.Binding
	TKey          key.Binding
	UKey          key.Binding
	WKey          key.Binding
	YKey          key.Binding
	YKeyCtrl      key.Binding
)

func Init() {
	Up = key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", lang.L("move up")))
	Down = key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", lang.L("move down")))
	Left = key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", lang.L("move left")))
	Right = key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", lang.L("move right")))
	ShiftLeft = key.NewBinding(
		key.WithKeys("shift+left", "H"),
		key.WithHelp("shift + ←/h", "move left step"))
	ShiftRight = key.NewBinding(
		key.WithKeys("shift+right", "L"),
		key.WithHelp("shift + →/l", "move right step"))
	Help = key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", lang.L("help")))
	Quit = key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", lang.L("quit")))
	Enter = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", lang.L("submit")))
	Space = key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp(lang.L("space"), lang.L("claim")))
	Esc = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", lang.L("back")))
	Filter = key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", lang.L("filter")))
	Tab = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", lang.L("next focus")))
	ShiftTab = key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp(lang.L("shift+tab"), lang.L("prev. focus")))
	Npcs = key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", lang.L("npcs")))
	Scripts = key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", lang.L("scripts")))
	GotoListStart = key.NewBinding(
		key.WithKeys("g", "home"),
		key.WithHelp("g/home", lang.L("goto list start")))
	GotoListEnd = key.NewBinding(
		key.WithKeys("G", "end"),
		key.WithHelp("G/end", lang.L("goto list end")))
	PrevPage = key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("page up", lang.L("previous page")))
	NextPage = key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("page down", lang.L("next page")))
	AKey = key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "a key"))
	CKey = key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "c key"))
	DKey = key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "d key"))
	EKey = key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "e key"))
	FKey = key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "f key"))
	HKey = key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "h key"))
	IKey = key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "i key"))
	LKey = key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "l key"))
	MKey = key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "m key"))
	NKey = key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "n key"))
	PKey = key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "p key"))
	RKey = key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "r key"))
	TKey = key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "t key"))
	UKey = key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "u key"))
	WKey = key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "w key"))
	YKey = key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "y key"))
	YKeyCtrl = key.NewBinding(
		key.WithKeys("ctrl+y"),
		key.WithHelp("ctrl+y", "ctrl+y key"))
}
