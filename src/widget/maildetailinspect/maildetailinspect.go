package maildetailinspect

import (
	"farental/art"
	"farental/core/data/api"
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

	mail        *api.MailBasicResponse
	attachments *[]api.MailAttachmentResponse
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	return w
}

func (w *Widget) Render() string {
	var b strings.Builder

	contentSize := w.GetContentSize()
	t := orvyn.GetTheme()

	if w.mail == nil {
		return ""
	}

	if !w.mail.HaveAttachments && w.mail.MoneyAmount == 0 {
		b.WriteString("No attachments")
	} else {
		if w.mail.MoneyAmount > 0 {
			fmt.Fprintf(&b, "%d %s", w.mail.MoneyAmount,
				t.Style(theme.HighlightTextStyleID).Render(string(art.CharGrynars)))
			b.WriteString("\n")
		}

		if len(*w.attachments) > 0 {
			b.WriteString("\n")
			for i, a := range *w.attachments {
				if i > 0 {
					b.WriteString("\n")
				}

				fmt.Fprintf(&b, "%c %dx %s",
					art.CharBullet, a.Amount, a.ItemName)
			}
		}

		if b.Len() > 0 && w.mail.IsAgainstPayment {
			b.WriteString("\n")
			b.WriteString(t.Style(ftheme.DimUnderlinedTextStyleID).
				Width(contentSize.Width).Render(""))
		}

		if w.mail.IsAgainstPayment {
			b.WriteString("\n\n")
			b.WriteString(lipgloss.NewStyle().Width(contentSize.Width).
				Render(fmt.Sprintf(
					lokyn.L("The sender asks you to pay %d %c to access the attachments."),
					w.mail.PaymentAmount, art.CharGrynars)))
		}
	}

	return w.GetStyle().
		Width(contentSize.Width).
		Height(contentSize.Height).
		Render(b.String())

}

func (w *Widget) UpdateData(mail *api.MailBasicResponse, attachments *[]api.MailAttachmentResponse) {
	if mail == nil {
		w.mail = nil

		return
	}

	w.mail = mail
	w.attachments = attachments
}
