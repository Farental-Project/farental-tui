package characterselection

import (
	"farental/art"
	"farental/internal/keybind"
	"farental/internal/orvyn"
	"farental/internal/orvyn/layout"
	"farental/style"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

const ID orvyn.ScreenID = "characterselection"

type Screen struct {
	orvyn.BaseScreen

	title *orvyn.StaticRenderable
}

func New() *Screen {
	s := new(Screen)

	s.title = orvyn.NewStaticRenderable(style.TitleStyle.Render(
		art.CreateASCIIArtTitle("wip screen"),
	))

	return s
}

func (s *Screen) OnEnter(_ interface{}) tea.Cmd {
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
		}
	}

	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return layout.NewCenterLayout(s.title)
}
