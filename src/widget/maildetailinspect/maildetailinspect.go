package maildetailinspect

import (
	"farental/art"
	"farental/core/data/api"
	"farental/style"
	"fmt"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
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

	if w.mail == nil {
		return ""
	}

	if !w.mail.HaveAttachments && w.mail.MoneyAmount == 0 {
		b.WriteString("No attachments")
	} else {
		if w.mail.MoneyAmount > 0 {
			b.WriteString(fmt.Sprintf("%d %s", w.mail.MoneyAmount,
				style.HighlightStyle.Render(string(art.CharGrynars))))
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
			b.WriteString(style.DimBottomBorderStyle.
				Width(w.contentSize.Width).Render(""))
		}

		if w.mail.IsAgainstPayment {
			b.WriteString("\n\n")
			b.WriteString(style.TextStyle.Width(w.contentSize.Width).
				Render(fmt.Sprintf(
					lokyn.L("The sender ask you to pay %d %c to access the attachments."),
					w.mail.PaymentAmount, art.CharGrynars)))
		}
	}

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
}

func (w *Widget) UpdateData(mail *api.MailBasicResponse, attachments *[]api.MailAttachmentResponse) {
	if mail == nil {
		w.mail = nil

		return
	}

	w.mail = mail
	w.attachments = attachments
}
