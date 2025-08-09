package maileditor

import (
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/layout"
	"farental/screen"
	"farental/style"
	"farental/widget/help"
	"farental/widget/mailattachmentlist"
	"farental/widget/mailattachmentselect"
	"farental/widget/maildetaileditor"
	"farental/widget/mailwriter"
	"farental/widget/statusmessage"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"net/http"
)

type Screen struct {
	orvyn.BaseScreen

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

	s.hideSelectAttachment()

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
				return orvyn.SwitchScreen(screen.IDMailBox)
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

	case mailattachmentlist.ShowAttachmentSelectMsg:
		s.showSelectAttachment()

	case mailattachmentlist.DeleteAttachmentMsg:
		s.detailEditor.RemoveAttachment(int(msg))

	case mailattachmentselect.HideAttachmentSelectMsg:
		s.hideSelectAttachment()

	case mailattachmentselect.SelectItemMsg:
		cmd := s.detailEditor.AddAttachment(maildetaileditor.ListItem{
			ItemName: msg.ItemName,
			Amount:   msg.Amount,
		})

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
	// Detect if it's a basic email or if there is attachments to it.
	if !s.detailEditor.HasAttachments() {
		mail := s.writer.GetMailBody()
		req := request.MailSend(mail)

		resp, err := helper.SendRequest(req)

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
}

func (s *Screen) hideSelectAttachment() {
	s.focusManager.ExitCurrentInput()

	s.attachmentSelect.SetActive(false)
	s.detailEditor.SetActive(true)

	s.focusManager.Focus(0)
}
