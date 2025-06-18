package itemdetail

import (
	"farental/core/data/api"
	"farental/internal/lang"
	"farental/style"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type layout struct {
	ItemDescribe string
	Stats        string
	Conditions   string
	Results      string
}

func (l *layout) buildLayout(data *api.StackResponse) {
	var b strings.Builder

	l.ItemDescribe = ""
	l.Stats = ""
	l.Conditions = ""
	l.Results = ""

	if data == nil {
		return
	}

	b.WriteString(style.TitleStyle.Render(data.Item.Name))
	b.WriteString("\n")
	if data.Item.IsUnique {
		b.WriteString(style.HighlightStyle.Render(lang.L("Unique")))
	}
	b.WriteString(style.DimTextStyle.Render(data.Item.Description))

	l.ItemDescribe = b.String()
	b.Reset()

	if data.Item.EquipmentSlot == nil {
		l.Stats = ""
		l.Conditions = ""
	} else {
		// Stats
		for i, s := range *data.Item.EquipmentStats {
			if i > 0 {
				b.WriteString("\n")
			}

			b.WriteString(fmt.Sprintf("%s : %d", s.Stat.Name, s.Value))
		}

		if b.Len() > 0 {
			l.Stats = b.String()
			b.Reset()
		}

		// Conditions
		for i, c := range *data.Item.Conditions {
			if i > 0 {
				b.WriteString("\n")
			}

			b.WriteString(fmt.Sprintf("* %s", c))
		}

		if b.Len() > 0 {
			l.Conditions = b.String()
			b.Reset()
		}
	}
}

type Model struct {
	data   *api.StackResponse
	layout *layout
}

func New(data *api.StackResponse) Model {
	m := Model{}

	m.UpdateData(data)

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	if m.data == nil {
		return ""
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		m.layout.ItemDescribe,
		style.DimBottomBorderStyle.Render("\n"),
		m.layout.Stats,
		style.DimBottomBorderStyle.Render("\n"),
		m.layout.Conditions,
		style.DimBottomBorderStyle.Render("\n"),
		m.layout.Results)
}

func (m *Model) UpdateData(data *api.StackResponse) {
	m.data = data
	m.layout = &layout{}
	m.layout.buildLayout(data)
}
