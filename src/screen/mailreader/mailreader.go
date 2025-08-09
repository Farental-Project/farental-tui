package mailreader

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/internal/orvyn/layout"
	"farental/widget/help"
	"farental/widget/mailcontentreader"
	"farental/widget/maildetailinspect"
	"farental/widget/statusmessage"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"net/http"
)

type Screen struct {
	orvyn.BaseScreen

	mail        *api.MailBasicResponse
	attachments []api.MailAttachmentResponse

	content   *mailcontentreader.Widget
	inspector *maildetailinspect.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	layoutContent *layout.HBoxFixedRatio
	layout        *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	s.content = mailcontentreader.New()
	s.inspector = maildetailinspect.New()
	s.statusMessage = statusmessage.New()
	s.help = help.New()

	s.layoutContent = layout.NewHBoxFixedRatioLayout(10, 1, 0,
		[]layout.FixedRatioRenderable{
			layout.NewFixedRatioRenderable(0.7, s.content),
			layout.NewFixedRatioRenderable(0.3, s.inspector),
		})

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(10, 4),
			0,
			[]orvyn.Renderable{
				s.layoutContent,
				s.statusMessage,
				orvyn.VGap,
				s.help,
			},
		),
	)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
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

func (s *Screen) OnExit() interface{} {
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
		s.statusMessage.SetMessage(lang.L("Payment sent !"),
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
		s.statusMessage.SetMessage(lang.L("Mail attachments transfered !"),
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
