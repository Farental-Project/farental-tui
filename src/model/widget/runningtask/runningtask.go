package runningtask

import (
	"farental/internal/context"
	"farental/internal/helper"
	"farental/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type Model struct {
	width int
	style lipgloss.Style
}

func New(width int) Model {
	return Model{
		width: width,
		style: style.TextStyle.Width(width).AlignHorizontal(lipgloss.Center),
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
		if context.RunningTask.RemainingTimeHours > 0 {
			b.WriteString("Running Task : " + context.RunningTask.Title +
				"\nRemaining time : " +
				helper.HoursDecFormat(context.RunningTask.RemainingTimeHours))
		} else {
			b.WriteString("Task completed! Waiting for claim!")
		}
	} else {
		b.WriteString("No running task")
	}

	return m.style.Render(b.String())
}
