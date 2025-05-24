package characterselection

import (
	"farental/internal"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	ctx           *internal.AppCtx
	width, height int
}

func New(ctx *internal.AppCtx) Model {
	return Model{ctx: ctx}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		s := msg.String()
		switch s {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) View() string {
	return "CHARACTER SELECTION CONTENT"
}
