package maildetailinspect

import (
	"farental/art"
	"farental/core/data/api"
	ftheme "farental/internal/theme"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"strings"
)

type Widget struct {
	orvyn.BaseWidget

	mail        *api.MailBasicResponse
	attachments *[]api.MailAttachmentResponse

	contentSize orvyn.Size
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	return w
}

func (w *Widget) Render() string {
	var b strings.Builder

	t := orvyn.GetTheme()

	if w.mail == nil {
		return ""
	}

	if !w.mail.HaveAttachments && w.mail.MoneyAmount == 0 {
		b.WriteString("No attachments")
	} else {
		if w.mail.MoneyAmount > 0 {
			b.WriteString(fmt.Sprintf("%d %s", w.mail.MoneyAmount,
				t.Style(theme.HighlightTextStyleID).Render(string(art.CharGrynars))))
			b.WriteString("\n")
		}

		if len(*w.attachments) > 0 {
			b.WriteString("\n")
			for i, a := range *w.attachments {
				if i > 0 {
					b.WriteString("\n")
				}

				b.WriteString(fmt.Sprintf("%c %dx %s",
					art.CharBullet, a.Amount, a.ItemName))
			}
		}

		if b.Len() > 0 && w.mail.IsAgainstPayment {
			b.WriteString("\n")
			b.WriteString(t.Style(ftheme.DimUnderlinedTextStyleID).
				Width(w.contentSize.Width).Render(""))
		}

		if w.mail.IsAgainstPayment {
			b.WriteString("\n\n")
			b.WriteString(lipgloss.NewStyle().Width(w.contentSize.Width).
				Render(fmt.Sprintf(
					lokyn.L("The sender ask you to pay %d %c to access the attachments."),
					w.mail.PaymentAmount, art.CharGrynars)))
		}
	}

	return t.Style(theme.BlurredWidgetStyleID).
		Width(w.contentSize.Width).
		Height(w.contentSize.Height).
		Render(b.String())

}

func (w *Widget) Resize(size orvyn.Size) {
	s := orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)

	w.BaseWidget.Resize(size)

	size.Width -= s.GetHorizontalFrameSize()
	size.Height -= s.GetVerticalFrameSize()

	w.contentSize = size
}

func (w *Widget) UpdateData(mail *api.MailBasicResponse, attachments *[]api.MailAttachmentResponse) {
	if mail == nil {
		w.mail = nil

		return
	}

	w.mail = mail
	w.attachments = attachments
}
