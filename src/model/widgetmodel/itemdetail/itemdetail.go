package itemdetail

import (
	"farental/core/data/api"
	"farental/internal/lang"
	"farental/style"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type layout struct {
	ItemDescribe string
	Stats        string
	Conditions   string
	Results      string

	width int
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

	if data.Item.EquipmentSlot != nil {
		// Stats
		for i, s := range *data.Item.EquipmentStats {
			if i > 0 {
				b.WriteString("\n")
			}

			b.WriteString(fmt.Sprintf("• %s : %d", s.Stat.Name, s.Value))
		}

		if b.Len() > 0 {
			l.Stats = l.renderContent(
				lang.L("Stats"),
				b.String())
			b.Reset()
		}

		// Conditions
		for i, c := range *data.Item.Conditions {
			if i > 0 {
				b.WriteString("\n")
			}

			b.WriteString(fmt.Sprintf("• %s", c))
		}

		if b.Len() > 0 {
			l.Conditions = l.renderContent(
				lang.L("Equip conditions"),
				b.String())
			b.Reset()
		}
	}

	if data.Item.Results != nil {
		for i, r := range *data.Item.Results {
			if i > 0 {
				b.WriteString("\n")
			}

			b.WriteString(fmt.Sprintf("• %s", r))
		}

		if b.Len() > 0 {
			l.Results = l.renderContent(
				lang.L("Results"),
				b.String())
			b.Reset()
		}
	}
}

func (l layout) renderContent(title, content string) string {
	return fmt.Sprintf("\n%s\n%s\n%s",
		style.DimBottomBorderStyle.
			Width(l.width).Render(" "),
		style.DimBottomBorderStyle.
			Width(l.width).Render(
			style.DimTextStyle.Render(title)),
		content)
}

type Model struct {
	data   *api.StackResponse
	layout *layout
}

func New(width int) Model {
	m := Model{}

	m.layout = &layout{}
	m.layout.width = width

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	if m.data == nil {
		return ""
	}

	b.WriteString(m.layout.ItemDescribe)
	b.WriteString(m.layout.Stats)
	b.WriteString(m.layout.Conditions)
	b.WriteString(m.layout.Results)

	return b.String()
}

func (m *Model) UpdateData(data *api.StackResponse) {
	m.data = data
	m.layout.buildLayout(data)
}

func (m *Model) SetWidth(width int) {
	m.layout.width = width
}

func (m *Model) GetWidth() int {
	return m.layout.width
}

func (m Model) GetDataItemID() uint {
	if m.data == nil {
		return 0
	}
	
	return m.data.ItemID
}
