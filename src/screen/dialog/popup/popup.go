package popup

import (
	"farental/art"
	"farental/internal/keybind"
	"farental/internal/orvyn"
	"farental/layout"
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type Option struct {
	Keybind key.Binding
	Text    string
	Value   uint
}

type Config struct {
	Message string
	Style   lipgloss.Style
	Options []Option
}

type Screen struct {
	config Config

	content *orvyn.SimpleRenderable
	options *orvyn.SimpleRenderable

	layout *layout.CenterLayout

	value uint
}

func New(config Config) *Screen {
	var b strings.Builder

	s := new(Screen)

	s.config = config

	b.WriteString(config.Style.Render(config.Message))
	b.WriteString("\n\n")

	s.content = orvyn.NewSimpleRenderable(b.String())
	s.content.SizeConstraint = true

	b.Reset()

	for i, o := range config.Options {
		if i > 0 {
			b.WriteString(fmt.Sprintf(" %c ", art.CharBullet))
		}

		b.WriteString(fmt.Sprintf("%s %s",
			style.NormalStyle.Render(o.Keybind.Help().Key),
			style.DimTextStyle.Render(o.Text)))
	}

	s.options = orvyn.NewSimpleRenderable(b.String())

	s.layout = layout.NewCenterLayout(
		layout.NewVBoxLayout(10,
			[]orvyn.Renderable{
				s.content,
				s.options,
			},
		),
	)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
	return nil
}

func (s *Screen) OnExit() interface{} {
	return s.value
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, keybind.Quit) {
			return tea.Quit
		}

		for _, o := range s.config.Options {
			if key.Matches(msg, o.Keybind) {
				s.value = o.Value
				return orvyn.CloseDialog()
			}
		}
	}

	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}
