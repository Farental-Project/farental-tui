package scripteditor

import (
	"farental/internal/orvyn"
	olist "farental/internal/orvyn/widget/list"
	"farental/layout"
	tea "github.com/charmbracelet/bubbletea"
)

type Screen struct {

	// ScriptInfoEditor
	// ScriptRulesList
	// - Complex list item

	// PileLayout :
	// ScriptRuleTypeSelect
	// ScriptRulesParamEditor
	// ScriptAbilitySelect

	list *olist.Widget[string]

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	s.list = olist.New(olist.SimpleListItemConstructor)

	s.layout = layout.NewCenterLayout(
		s.list,
	)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
	data := []string{
		"Hello",
		"Coucou",
		"Salut",
		"Fine",
		"Ta race",
		"Hello",
		"Coucou",
		"Salut",
		"Fine",
		"Ta race",
		"Hello",
		"Coucou",
		"Salut",
		"Fine",
		"Ta race",
		"Hello",
		"Coucou",
		"Salut",
		"Fine",
		"Ta race",
		"Hello",
		"Coucou",
		"Salut",
		"Fine",
		"Ta race",
		"Hello",
		"Coucou",
		"Salut",
		"Fine",
		"Ta race",
	}

	s.list.SetItems(data)

	return nil
}

func (s *Screen) OnExit() interface{} {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	cmd := s.list.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}
