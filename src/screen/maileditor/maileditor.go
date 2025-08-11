package maileditor

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/layout"
	"farental/screen"
	"farental/screen/dialog/popup"
	"farental/style"
	"farental/widget/help"
	"farental/widget/mailattachmentlist"
	"farental/widget/mailattachmentselect"
	"farental/widget/maildetaileditor"
	"farental/widget/mailwriter"
	"farental/widget/statusmessage"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-resty/resty/v2"
	"net/http"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	writer           *mailwriter.Widget
	detailEditor     *maildetaileditor.Widget
	attachmentSelect *mailattachmentselect.Widget
	statusMessage    *statusmessage.Widget
	help             *help.Widget

	focusManager *orvyn.FocusManager

	detailLayout *layout.PileLayout
	editorLayout *layout.HBoxFixedRatio
	layout       *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	s.title = orvyn.NewSimpleRenderable(lang.L("New mail"))
	s.title.Style = style.TitleStyle

	s.writer = mailwriter.New()
	s.detailEditor = maildetaileditor.New()
	s.attachmentSelect = mailattachmentselect.New()
	s.attachmentSelect.SetActive(false)
	s.statusMessage = statusmessage.New()
	s.help = help.New()

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.writer)
	s.focusManager.Add(s.detailEditor)
	s.focusManager.Add(s.attachmentSelect)

	s.detailLayout = layout.NewPileLayout(
		[]orvyn.Renderable{
			s.detailEditor,
			s.attachmentSelect,
		},
	)

	s.editorLayout = layout.NewHBoxFixedRatioLayout(
		0, 1, 0,
		[]layout.FixedRatioRenderable{
			layout.NewFixedRatioRenderable(0.7, s.writer),
			layout.NewFixedRatioRenderable(0.3, s.detailLayout),
		},
	)

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(
			orvyn.NewSize(10, 4),
			2,
			[]orvyn.Renderable{
				s.title,
				orvyn.VGap,
				s.editorLayout,
				s.statusMessage,
				s.help,
			},
		),
	)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
	s.writer.Init()
	s.detailEditor.Init()
	s.attachmentSelect.Init()

	s.hideSelectAttachment()

	s.focusManager.ExitCurrentInput()
	s.focusManager.Focus(0)

	s.statusMessage.Reset()

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
			if !s.writer.IsInputting() && !s.detailEditor.IsInputting() &&
				!s.attachmentSelect.IsInputting() {

				options := []popup.Option{
					{
						Keybind: keybind.YKey,
						Text:    lang.L("Yes"),
						Value:   1,
					},
					{
						Keybind: keybind.NKey,
						Text:    lang.L("No"),
						Value:   2,
					},
				}

				config := popup.Config{
					Message: "Are you sure you want to exit the mail and loose your work ?",
					Style:   style.TitleStyle,
					Options: options,
				}

				orvyn.OpenDialog("quitConfirm", popup.New(config), nil)
				return nil
			}

		case key.Matches(msg, keybind.Enter):
			if !s.writer.IsInputting() && !s.detailEditor.IsInputting() &&
				!s.attachmentSelect.IsInputting() {
				if s.submit() {
					s.OnEnter(nil)
					s.statusMessage.SetMessage(lang.L("Mail successfully sent !"), statusmessage.SuccessMessage)
				}

				return nil
			}

		}

	case orvyn.DialogExitMsg:
		switch msg.DialogID {
		case "quitConfirm":
			val := msg.Param.(uint)
			switch val {
			case 1:
				return orvyn.SwitchScreen(screen.IDMailBox)
			default:
				return nil
			}
		}

	case mailattachmentlist.ShowAttachmentSelectMsg:
		s.showSelectAttachment()

	case mailattachmentlist.DeleteAttachmentMsg:
		s.detailEditor.RemoveAttachment(int(msg))

	case mailattachmentselect.HideAttachmentSelectMsg:
		s.hideSelectAttachment()

	case mailattachmentselect.SelectItemMsg:
		cmd, err := s.detailEditor.AddAttachment(maildetaileditor.ListItem{
			StackID:  msg.StackID,
			ItemID:   msg.ItemID,
			ItemName: msg.ItemName,
			Amount:   msg.Amount,
		})

		if err != nil {
			s.statusMessage.SetError(err)
		}

		s.hideSelectAttachment()

		return cmd
	}

	cmd := s.focusManager.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) submit() bool {
	var attachments []api.MailAttachment
	var req *resty.Request
	var resp *resty.Response
	var err error

	// Detect if it's a basic email or if there is attachments to it.
	if !s.detailEditor.HasAttachments() {
		mail := s.writer.GetMailBody()
		req = request.MailSend(mail)

		resp, err = helper.SendRequest(req)

		if err != nil {
			s.statusMessage.SetError(err)
			return false
		}

		if resp.StatusCode() == http.StatusOK {
			return true
		}
	} else {
		mail := api.MailWithAttachmentsBody{
			MailSendBody:     s.writer.GetMailBody(),
			IsAgainstPayment: false,
			MoneyAmount:      s.detailEditor.GetAttachedMoneyAmount(),
		}

		mail.PaymentAmount = s.detailEditor.GetPaymentAmount()

		if mail.PaymentAmount > 0 {
			mail.IsAgainstPayment = true
		}

		attachments = make([]api.MailAttachment, 0)

		for _, a := range s.detailEditor.GetAttachments() {
			attachment := api.MailAttachment{
				ItemID: a.ItemID,
				Amount: a.Amount,
			}

			attachments = append(attachments, attachment)
		}

		mail.Items = attachments

		req = request.MailSendWithAttachments(mail)

		resp, err = helper.SendRequest(req)

		if err != nil {
			s.statusMessage.SetError(err)
			return false
		}

		if resp.StatusCode() == http.StatusOK {
			return true
		}
	}

	return false
}

func (s *Screen) showSelectAttachment() {
	s.focusManager.ExitCurrentInput()

	s.attachmentSelect.SetActive(true)
	s.detailEditor.SetActive(false)

	s.focusManager.Focus(2)
	s.focusManager.ForceInput(2)

	// Load data with filter list
	currentAttachments := s.detailEditor.GetAttachments()

	filterItems := make([]mailattachmentselect.ListItem, 0)

	for _, i := range currentAttachments {
		item := mailattachmentselect.ListItem{
			Stack:  api.StackResponse{ID: i.StackID},
			Amount: i.Amount,
		}

		filterItems = append(filterItems, item)
	}

	s.attachmentSelect.LoadData(filterItems)
}

func (s *Screen) hideSelectAttachment() {
	s.focusManager.ExitCurrentInput()

	s.attachmentSelect.SetActive(false)
	s.detailEditor.SetActive(true)

	s.focusManager.Focus(1)
	s.focusManager.ForceInput(1)
	s.detailEditor.SetFocusOnAttachmentList()
}
