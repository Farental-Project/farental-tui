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
			authorName := w.data.AuthorName
			if authorName == "" {
				authorName = lokyn.L("deleted user")
			}
			right.WriteString(ds.Render(fmt.Sprintf(lokyn.L("Author : %s"), authorName)))
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

// GetMinSize measures an actual render. The height depends on whether a script is
// selected and on how the description wraps, so the previous "description height
// plus five" guess was three rows too tall with no script selected - rows the
// layout reserved and nothing drew in.
func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(30, lipgloss.Height(w.Render()))
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(45, lipgloss.Height(w.Render()))
}

func (w *Widget) UpdateData() {
	activeScript, err := helper.Fetch[api.ScriptBasicResponse](request.ScriptGetActive())

	if err != nil {
		w.data.ID = nil
		return
	}

	w.data = *activeScript
}
