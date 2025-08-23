package mailcontentreader

import (
	"farental/art"
	"farental/core/data/api"
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"strings"
)

type Widget struct {
	orvyn.BaseWidget

	from     string
	subject  string
	read     bool
	viewport viewport.Model

	contentSize orvyn.Size

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

	readStatus := ""

	if !w.read {
		readStatus = string(art.CharBullet)
	}

	from := fmt.Sprintf("%s : %s",
		style.TitleStyle.Render(lokyn.L("From")),
		w.from)
	subject := fmt.Sprintf("%s : %s",
		style.TitleStyle.Render(lokyn.L("Subject")),
		w.subject)

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		style.TextStyle.Width(w.contentSize.Width-2).
			AlignHorizontal(lipgloss.Left).
			Render(from),
		style.SpecialHighlightStyle.Width(1).
			AlignHorizontal(lipgloss.Right).
			Render(readStatus)))
	b.WriteString("\n")
	b.WriteString(
		style.DimBottomBorderStyle.
			Width(w.contentSize.Width).Render(subject),
	)
	b.WriteString("\n")
	b.WriteString(w.viewport.View())

	return style.BlurredStyle.
		Width(w.contentSize.Width).
		Height(w.contentSize.Height).
		Render(b.String())
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	size.Width -= style.BlurredStyle.GetHorizontalFrameSize()
	size.Height -= style.BlurredStyle.GetVerticalFrameSize()

	w.contentSize = size
	w.viewport.Width = size.Width
	w.viewport.Height = size.Height - w.headerHeight
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
