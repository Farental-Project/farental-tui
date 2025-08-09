package maileditor

import (
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/layout"
	"farental/style"
	"farental/widget/help"
	"farental/widget/maildetaileditor"
	"farental/widget/mailwriter"
	"farental/widget/statusmessage"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type Screen struct {
	orvyn.BaseScreen

	title *orvyn.SimpleRenderable

	writer        *mailwriter.Widget
	detailEditor  *maildetaileditor.Widget
	statusMessage *statusmessage.Widget
	help          *help.Widget

	focusManager *orvyn.FocusManager

	editorLayout *layout.HBoxFixedRatio
	layout       *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	s.title = orvyn.NewSimpleRenderable(lang.L("New mail"))
	s.title.Style = style.TitleStyle

	s.writer = mailwriter.New()
	s.detailEditor = maildetaileditor.New()
	s.statusMessage = statusmessage.New()
	s.help = help.New()

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.writer)
	s.focusManager.Add(s.detailEditor)

	s.editorLayout = layout.NewHBoxFixedRatioLayout(
		0, 1, 0,
		[]layout.FixedRatioRenderable{
			layout.NewFixedRatioRenderable(0.7, s.writer),
			layout.NewFixedRatioRenderable(0.3, s.detailEditor),
		},
	)

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(
			orvyn.NewSize(10, 4),
			2,
			[]orvyn.Renderable{
				s.title,
				orvyn.VGap,
				s.editorLayout,
				s.statusMessage,
				s.help,
			},
		),
	)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
	s.writer.Init()
	s.detailEditor.Init()

	s.focusManager.Focus(0)

	return nil
}

func (s *Screen) OnExit() interface{} {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return tea.Quit

		case key.Matches(msg, keybind.Esc):
			if !s.writer.IsInputting() && !s.detailEditor.IsInputting() {
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
