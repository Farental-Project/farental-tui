package scriptexplorerlistitem

import (
	"farental/core/data/api"
	"farental/internal/context"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data api.ScriptBasicResponse
}

func Constructor(data api.ScriptBasicResponse) list.ListItem[api.ScriptBasicResponse] {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.data = data

	w.OnBlur()

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 4

	w.BaseWidget.Resize(size)
}

func (w *Widget) UpdateData(data api.ScriptBasicResponse) {
	w.data = data
}

func (w *Widget) GetData() api.ScriptBasicResponse {
	return w.data
}

func (w *Widget) Render() string {
	var s lipgloss.Style
	var left strings.Builder
	var right strings.Builder
	var width int

	contentSize := w.GetContentSize()
	width = contentSize.Width

	s = w.GetStyle()
	t := orvyn.GetTheme()
	ds := t.Style(theme.DimTextStyleID)
	ns := lipgloss.NewStyle()

	left.WriteString(t.Style(theme.TitleStyleID).Render(w.data.Name))
	left.WriteString("\n")
	left.WriteString(ds.Render(w.data.Description))

	if !w.data.IsEditable {
		right.WriteString(ds.Render(fmt.Sprintf(lokyn.L("Author : %s"), w.data.AuthorName)))
	} else {
		if w.data.IsPrivate {
			right.WriteString(ds.Render(lokyn.L("Private")))
		} else {
			right.WriteString(ds.Render(lokyn.L("Public")))
		}
	}

	if w.data.ID == *context.CharacterInfo.ScriptID {
		right.WriteString("\n")
		right.WriteString(t.Style(theme.TitleStyleID).Render(lokyn.L("Active")))
	}

	width1, width2 := orvyn.DivideSizeFull(width)

	tui := s.Width(width).Height(contentSize.Height).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			ns.Width(width1).
				AlignHorizontal(lipgloss.Left).
				Render(left.String()),
			ns.Width(width2).
				AlignHorizontal(lipgloss.Right).
				Render(right.String())))

	return tui
}

func (w *Widget) FilterValue() string {
	var b strings.Builder

	b.WriteString(w.data.Name)
	b.WriteString(" ")
	b.WriteString(w.data.Description)

	return b.String()
}
