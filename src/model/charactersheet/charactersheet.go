package charactersheet

import (
	"farental/core/data/api"
	"farental/internal/context"
	"farental/internal/keybind"
	"farental/model"
	"farental/model/widget/charactervitalinfo"
	"farental/style"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Data   api.CharacterInfoResponse
	ErrMsg error

	CharacterVitalInfo charactervitalinfo.Model
}

func New() Model {
	m := Model{
		CharacterVitalInfo: charactervitalinfo.New(style.LayoutWidth),
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return model.InitCmd
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	defer context.ContentManager.UpdateCurrentContent(m)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return m, tea.Quit
		case key.Matches(msg, keybind.Esc):
			return context.ContentManager.
				SwitchContent(m, model.ContentGameDashboard)
		}
	case model.InitMsg:
		context.KeymapManager.SwitchContext(model.ContextCharacterSheet)
	}

	context.ContentManager.Update(msg)

	return m, nil
}

func (m Model) View() string {
	tui := lipgloss.JoinVertical(lipgloss.Center,
		style.ContainerStyle.Render(m.CharacterVitalInfo.View()),
		context.KeymapManager.View(style.LayoutWidth))

	return lipgloss.Place(
		context.ContentManager.ScreenWidth,
		context.ContentManager.ScreenHeight,
		lipgloss.Center,
		lipgloss.Center,
		tui)
}
