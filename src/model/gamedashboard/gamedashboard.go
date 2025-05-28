package gamedashboard

import (
	"farental/core/data/api"
	"farental/internal/context"
	"farental/model"
	"farental/model/widget/charactervitalinfo"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	styleDashboard = lipgloss.NewStyle().Width(100).AlignHorizontal(lipgloss.Center)
)

type Model struct {
	CharacterVitalInfo charactervitalinfo.Model
}

func New() Model {
	return Model{
		CharacterVitalInfo: charactervitalinfo.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return model.InitCmd
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case model.InitMsg:
		// Test data
		m.CharacterVitalInfo.UpdateData(&api.CharacterInfoResponse{
			ID:        1,
			FirstName: "JeanJean",
			LastName:  "Poulos",
			RaceName:  "Garnoth",
			Power:     3632,
			Stats: []api.CharacterStatResponse{
				{
					Code:     "hp",
					Value:    36,
					MaxValue: 100,
				},
				{
					Code:     "mp",
					Value:    56,
					MaxValue: 120,
				},
			},
			Location: api.LocationResponse{},
		})
	}

	context.ContentManager.Update(msg)

	return m, nil
}

func (m Model) View() string {
	return m.CharacterVitalInfo.View()
}
