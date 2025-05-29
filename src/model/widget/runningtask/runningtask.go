package runningtask

import (
	"farental/internal/context"
	"farental/internal/helper"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type Model struct {
	width int
}

func New(width int) Model {
	return Model{
		width: width,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	if context.RunningTask != nil {
		b.WriteString("Running Task : " + context.RunningTask.Title +
			" Remaining time : " +
			helper.HoursDecFormat(context.RunningTask.RemainingTimeHours))
	} else {
		b.WriteString("No running task found")
	}

	return ""
}
