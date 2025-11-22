package mailreader

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/widget/help"
	"farental/widget/mailcontentreader"
	"farental/widget/maildetailinspect"
	"net/http"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
)

type Screen struct {
	mail        *api.MailBasicResponse
	attachments []api.MailAttachmentResponse

	title     *orvyn.SimpleRenderable
	content   *mailcontentreader.Widget
	inspector *maildetailinspect.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	layoutContent *layout.HBoxFixedRatio
	layout        *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	s.title = orvyn.NewSimpleRenderable(lokyn.L("Read mail"))
	s.title.Style = orvyn.GetTheme().Style(theme.TitleStyleID)

	s.content = mailcontentreader.New()
	s.inspector = maildetailinspect.New()
	s.statusMessage = statusmessage.New()
	s.help = help.New()

	layoutElements := []layout.FixedRatioRenderable{
		layout.NewFixedRatioRenderable(0.7, s.content),
		layout.NewFixedRatioRenderable(0.3, s.inspector),
	}

	s.layoutContent = layout.NewHBoxFixedRatioLayout(10, 1, 0, layoutElements...)

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(10, 4),
			2,
			s.title,
			orvyn.VGap,
			s.layoutContent,
			s.statusMessage,
			s.help,
		),
	)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextMailReader)

	s.statusMessage.Reset()

	s.mail = nil

	mail, ok := i.(*api.MailBasicResponse)

	if !ok {
		return orvyn.SwitchToPreviousScreen()
	}

	s.mail = mail

	s.content.UpdateData(mail)

	s.updateAttachments()
	s.updateKeymap()
	s.changeReadStatus(true)

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s.statusMessage.Reset()

		switch {
		case key.Matches(msg, keybind.Quit):
			return tea.Quit

		case key.Matches(msg, keybind.Esc):
			return orvyn.SwitchToPreviousScreen()

		case key.Matches(msg, keybind.Help):
			bubblehelp.ShowAll = !bubblehelp.ShowAll

			return nil

		case key.Matches(msg, keybind.PKey):
			if bubblehelp.IsKeybindVisible(keybind.PKey) {
				s.payMail()

				return nil
			}

		case key.Matches(msg, keybind.TKey):
			if bubblehelp.IsKeybindVisible(keybind.TKey) {
				s.transferAttachments()

				return nil
			}

		case key.Matches(msg, keybind.RKey):
			s.changeReadStatus(!s.mail.IsRead)

			return nil
		}
	}

	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) updateData() {
	req := request.MailGetOne(s.mail.ID)

	resp, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	mail := resp.Result().(*api.MailBasicResponse)

	s.mail = mail

	s.content.UpdateData(mail)
	s.updateAttachments()
}

func (s *Screen) updateAttachments() {
	if s.mail == nil {
		return
	}

	req := request.MailGetAttachments(s.mail.ID)

	resp, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.attachments = *resp.Result().(*[]api.MailAttachmentResponse)

	s.inspector.UpdateData(s.mail, &s.attachments)
}

func (s *Screen) updateKeymap() {
	if s.mail == nil {
		return
	}

	if !s.mail.IsAgainstPayment {
		bubblehelp.SetKeybindVisible(keybind.PKey, false)
	}

	if s.mail.IsAgainstPayment || (len(s.attachments) == 0 && s.mail.MoneyAmount == 0) {
		bubblehelp.SetKeybindVisible(keybind.TKey, false)
	}

}

func (s *Screen) payMail() {
	req := request.MailPay(s.mail.ID)

	resp, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	if resp.StatusCode() == http.StatusOK {
		s.statusMessage.SetMessage(lokyn.L("Payment sent !"),
			statusmessage.SuccessMessage)
		s.updateData()
		s.updateKeymap()
	}
}

func (s *Screen) transferAttachments() {
	req := request.MailTransferAttachments(s.mail.ID)

	resp, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	if resp.StatusCode() == http.StatusOK {
		s.statusMessage.SetMessage(lokyn.L("Mail attachments transfered !"),
			statusmessage.SuccessMessage)
		s.updateData()
		s.updateKeymap()
	}
}

func (s *Screen) changeReadStatus(read bool) {
	req := request.MailSetRead(s.mail.ID, read)

	resp, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	if resp.StatusCode() == http.StatusOK {
		s.updateData()
		s.updateKeymap()
	}
}
