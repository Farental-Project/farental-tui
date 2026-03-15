package characteractivescript

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	ftheme "farental/internal/theme"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
)

type Widget struct {
	orvyn.BaseWidget

	title string

	data api.ScriptBasicResponse
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.title = lokyn.L("Active script")

	return w
}

func (w *Widget) Render() string {
	var left, right strings.Builder
	var content string

	t := orvyn.GetTheme()
	ds := t.Style(theme.DimTextStyleID)
	ns := lipgloss.NewStyle()

	size := w.GetContentSize()
	width := size.Width

	if len(w.data.ID) == 0 {
		content = ds.Render(lokyn.L("No active script selected"))
	} else {
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

		width1, width2 := orvyn.DivideSizeFull(width)

		summary := lipgloss.JoinHorizontal(lipgloss.Top,
			ns.Width(width1).
				AlignHorizontal(lipgloss.Left).
				Render(left.String()),
			ns.Width(width2).
				AlignHorizontal(lipgloss.Right).
				Render(right.String()))

		content = lipgloss.JoinVertical(lipgloss.Left,
			t.Style(ftheme.DimUnderlinedTextStyleID).
				Width(size.Width).
				Render(w.title),
			summary)
	}

	return w.GetStyle().
		Width(size.Width).Render(content)
}

func (w *Widget) UpdateData() {
	resp, err := helper.SendRequest(request.ScriptGetActive())

	if err != nil {
		w.data.ID = nil
		return
	}

	activeScript, ok := resp.Result().(*api.ScriptBasicResponse)

	if !ok {
		w.data.ID = nil
		return
	}

	w.data = *activeScript
}
