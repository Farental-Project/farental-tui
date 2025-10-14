package mailcontentreader

import (
	"farental/art"
	"farental/core/data/api"
	ftheme "farental/internal/theme"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
)

type Widget struct {
	orvyn.BaseWidget

	from     string
	subject  string
	read     bool
	viewport viewport.Model

	headerHeight int
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.viewport = viewport.New(0, 0)

	w.headerHeight = 3 // 2 field + border

	return w
}

func (w *Widget) Render() string {
	var b strings.Builder

	contentSize := w.GetContentSize()
	t := orvyn.GetTheme()
	ts := t.Style(theme.TitleStyleID)
	ns := lipgloss.NewStyle()

	readStatus := ""

	if !w.read {
		readStatus = string(art.CharBullet)
	}

	from := fmt.Sprintf("%s : %s",
		ts.Render(lokyn.L("From")),
		w.from)
	subject := fmt.Sprintf("%s : %s",
		ts.Render(lokyn.L("Subject")),
		w.subject)

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		ns.Width(contentSize.Width-2).
			AlignHorizontal(lipgloss.Left).
			Render(from),
		t.Style(theme.HighlightTextStyleID).Width(1).
			AlignHorizontal(lipgloss.Right).
			Render(readStatus)))
	b.WriteString("\n")
	b.WriteString(
		t.Style(ftheme.DimUnderlinedTextStyleID).
			Width(contentSize.Width).Render(subject),
	)
	b.WriteString("\n")
	b.WriteString(w.viewport.View())

	return t.Style(theme.BlurredWidgetStyleID).
		Width(contentSize.Width).
		Height(contentSize.Height).
		Render(b.String())
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	contentSize := w.GetContentSize()

	w.viewport.Width = contentSize.Width
	w.viewport.Height = contentSize.Height - w.headerHeight
}

func (w *Widget) UpdateData(mail *api.MailBasicResponse) {
	if mail == nil {
		w.from = ""
		w.subject = ""
		w.read = true
		w.viewport.SetContent("")

		return
	}

	w.from = mail.SenderName
	w.subject = mail.Subject
	w.read = mail.IsRead
	w.viewport.SetContent(mail.Content)
}
