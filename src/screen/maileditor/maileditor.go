package maileditor

import (
	"farental/internal/keybind"
	"farental/internal/orvyn"
	"farental/internal/orvyn/layout"
	"farental/widget/mailwriter"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type Screen struct {
	orvyn.BaseScreen

	writer *mailwriter.Widget

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	s.writer = mailwriter.New()

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.writer)

	s.layout = layout.NewCenterLayout(s.writer)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
	s.writer.Init()

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
			if !s.writer.IsInputting() {
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
