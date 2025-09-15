package scripteditor

import (
	"farental/core/data/api"
	"farental/internal/keybind"
	"farental/widget/help"
	"farental/widget/scriptinfoinput"
	"farental/widget/scriptrulelist"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
)

type Screen struct {
	title         *orvyn.SimpleRenderable
	scriptInfo    *scriptinfoinput.Widget
	list          *scriptrulelist.Widget
	statusMessage *statusmessage.Widget
	help          *help.Widget

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout

	new bool
}

func New() *Screen {
	s := new(Screen)

	s.title = orvyn.NewSimpleRenderable(lokyn.L("Script editor"))
	s.title.Style = orvyn.GetTheme().Style(theme.TitleStyleID)

	s.scriptInfo = scriptinfoinput.New()

	s.list = scriptrulelist.New()
	s.list.SetFilterable(false)

	s.statusMessage = statusmessage.New()
	s.help = help.New()

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.scriptInfo)
	s.focusManager.Add(s.list)

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(
			orvyn.NewSize(10, 4),
			2,
			[]orvyn.Renderable{
				s.title,
				orvyn.VGap,
				layout.NewHBoxFixedRatioLayout(
					0, 1, 1,
					[]layout.FixedRatioRenderable{
						layout.NewFixedRatioRenderable(0.2, s.scriptInfo),
						layout.NewFixedRatioRenderable(0.8, s.list),
					},
				),
				s.statusMessage,
				s.help,
			},
		),
	)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	script, ok := i.(*api.ScriptBasicResponse)

	if !ok || script == nil {
		s.new = true
	} else {

	}

	s.focusManager.FocusFirst()

	s.list.Init()

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if m, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(m, keybind.Esc):
			if !s.focusManager.IsInputting() && !s.list.IsInputting() {
				return orvyn.SwitchToPreviousScreen()
			}
		}
	}

	cmd := s.focusManager.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}
