package activitylistitem

import (
	"farental/core/data/api"
	"farental/internal/helper"
	"farental/internal/keybind"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
	"strings"
)

type DurationData struct {
	api.DurationResponse
}

func (d DurationData) RenderValue() string {
	return helper.HoursDecFormat(d.Duration)
}

type Data struct {
	api.ActivityResponse
	DurationIndex int
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	style lipgloss.Style

	data *Data

	contentSize orvyn.Size
}

func Constructor(data *Data) list.IListItem {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.data = data

	w.OnBlur()

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch {
		case key.Matches(msg, keybind.Left):
			w.data.DurationIndex--

			if w.data.DurationIndex < 0 {
				w.data.DurationIndex = 0
			}

		case key.Matches(msg, keybind.Right):
			w.data.DurationIndex++

			length := len(w.data.Duration.Durations) - 1

			if w.data.DurationIndex > length {
				w.data.DurationIndex = length
			}

		}
	}

	return nil
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 6

	w.BaseWidget.Resize(size)

	size.Width -= w.style.GetHorizontalFrameSize()
	size.Height -= w.style.GetVerticalFrameSize()

	w.contentSize = size
}

func (w *Widget) Render() string {
	var s lipgloss.Style
	var left strings.Builder
	var right strings.Builder
	var width int

	width = w.contentSize.Width

	s = w.style
	t := orvyn.GetTheme()
	ds := t.Style(theme.DimTextStyleID)
	hs := t.Style(theme.HighlightTextStyleID)
	ns := lipgloss.NewStyle()

	left.WriteString(hs.Render(w.data.Name))
	left.WriteString("\n")
	left.WriteString(ds.Render(w.data.Description))

	right.WriteString(ds.Render(w.data.Skill.Name))
	right.WriteString("\n\n\n")

	if len(w.data.Duration.Durations) > 0 {
		right.WriteString(hs.Render("< "))
		right.WriteString(t.Style(theme.NormalTextStyleID).
			Bold(true).Render(helper.HoursDecFormat(
			w.data.Duration.Durations[w.data.DurationIndex].Duration)))
		right.WriteString(hs.Render(" >"))
	} else {
		right.WriteString(t.Style(theme.NormalTextStyleID).
			Render(helper.HoursDecFormat(
				w.data.Duration.Durations[0].Duration)))
	}

	tui := s.Width(width).Height(w.contentSize.Height).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			ns.Width(width/2).
				AlignHorizontal(lipgloss.Left).
				Render(left.String()),
			ns.Width(width/2).
				AlignHorizontal(lipgloss.Right).
				Render(right.String())))

	return tui
}

func (w *Widget) OnFocus() {
	w.style = orvyn.GetTheme().Style(theme.FocusedWidgetStyleID)
}

func (w *Widget) OnBlur() {
	w.style = orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)
}

func (w *Widget) OnEnterInput() {}

func (w *Widget) OnExitInput() {}

func (w *Widget) FilterValue() string {
	var b strings.Builder

	b.WriteString(w.data.Name)
	b.WriteString(" ")
	b.WriteString(w.data.Description)
	b.WriteString(" ")
	b.WriteString(w.data.Skill.Name)

	return b.String()
}
