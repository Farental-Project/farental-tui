package runningtask

import (
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/lang"
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"time"
)

type Model struct {
	width int
	style lipgloss.Style

	spinner spinner.Model
}

func New(width int) Model {
	m := Model{
		width: width,
		style: style.TextStyle.Width(width).AlignHorizontal(lipgloss.Center),
	}

	// Characters
	// "█", "▓", "▒", "░"},

	m.spinner = spinner.New()
	m.spinner.Style = style.TitleStyle
	m.spinner.Spinner = spinner.Spinner{
		Frames: []string{
			"░░░░░░░░░░",
			"▒░░░░░░░░░",
			"▓▒░░░░░░░░",
			"█▓▒░░░░░░░",
			"▓█▓▒░░░░░░",
			"▒▓█▓▒░░░░░",
			"░▒▓█▓▒░░░░",
			"░░▒▓█▓▒░░░",
			"░░░▒▓█▓▒░░",
			"░░░░▒▓█▓▒░",
			"░░░░░▒▓█▓▒",
			"░░░░░░▒▓█▓",
			"░░░░░░░▒▓█",
			"░░░░░░░░▒▓",
			"░░░░░░░░░▒",
		},
		FPS: time.Second / 9, //nolint:mnd
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	if context.RunningTask != nil {
		b.WriteString(style.TitleStyle.Render(context.RunningTask.Title))
		b.WriteString("\n")

		if context.RunningTask.RemainingTimeHours > 0 {
			b.WriteString(fmt.Sprintf("%s : %s", lang.L("Remaining time"),
				helper.HoursDecFormat(context.RunningTask.RemainingTimeHours)))
			b.WriteString("\n")
			b.WriteString(style.ContainerStyle.Render(m.spinner.View()))
		} else {
			b.WriteString(lang.L("Completed! Waiting for claim!"))
		}
	} else {
		b.WriteString(style.NeutralDimTextStyle.Render(lang.L("No running task")))
	}

	return m.style.Render(b.String())
}
