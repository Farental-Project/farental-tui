package mailreader

import (
	"farental/art"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/model"
	"farental/model/mailbox"
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"net/http"
	"strings"
)

type Model struct {
	Mail        *api.MailBasicResponse
	Attachments []api.MailAttachmentResponse

	ErrMsg error

	VPContent viewport.Model

	widthLeft, widthRight int
}

func New() Model {
	m := Model{}

	m.widthLeft = style.LayoutWidth - 30
	m.widthRight = 24

	m.VPContent = viewport.New(m.widthLeft, 20)

	return m
}

func (m Model) Init() tea.Cmd {
	return model.InitCmd
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	defer context.ContentManager.UpdateCurrentContent(m)

	switch msg := msg.(type) {
	case model.InitMsg:
		bubblehelp.SwitchContext(keybind.ContextMailReader)

		content := context.ContentManager.GetContent(model.ContentMailbox)

		mailbox := content.(mailbox.Model)

		m.Mail = mailbox.SelectedMail

		m.VPContent.SetContent(m.Mail.Content)

		m.updateAttachments()

		m.updateKeymap()

		m.changeReadStatus(true)

	case tea.KeyMsg:
		m.ErrMsg = nil
		switch {
		case key.Matches(msg, keybind.Quit):
			return m, tea.Quit

		case key.Matches(msg, keybind.Help):
			bubblehelp.ShowAll = !bubblehelp.ShowAll

			return m, nil

		case key.Matches(msg, keybind.Esc):
			return context.ContentManager.Back(m)

		case key.Matches(msg, keybind.PKey):
			if bubblehelp.IsKeybindVisible(keybind.PKey) {
				m.payMail()

				return m, nil
			}

		case key.Matches(msg, keybind.TKey):
			if bubblehelp.IsKeybindVisible(keybind.TKey) {
				m.transferAttachments()

				return m, nil
			}

		case key.Matches(msg, keybind.RKey):
			m.changeReadStatus(!m.Mail.IsRead)
		}
	}

	context.ContentManager.Update(msg)

	return m, nil
}

func (m Model) View() string {
	var left, right, tui strings.Builder
	var readStatus string
	var from string

	if m.Mail == nil {
		return ""
	}

	readStatus = ""

	if !m.Mail.IsRead {
		readStatus = string(art.CharBullet)
	}

	from = fmt.Sprintf("%s : %s",
		style.TitleStyle.Render(lang.L("From")),
		m.Mail.SenderName)

	left.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		style.TextStyle.Width(m.widthLeft-2).
			AlignHorizontal(lipgloss.Left).
			Render(from),
		style.TextStyle.Width(1).
			AlignHorizontal(lipgloss.Right).
			Render(readStatus)))
	left.WriteString("\n")
	left.WriteString(style.DimBottomBorderStyle.
		Width(m.widthLeft).
		Render(fmt.Sprintf("%s : %s",
			style.TitleStyle.Render(lang.L("Subject")),
			m.Mail.Subject)))
	left.WriteString("\n")
	left.WriteString(m.VPContent.View())

	if !m.Mail.HaveAttachments && m.Mail.MoneyAmount == 0 {
		right.WriteString("No attachments")
	} else {
		if m.Mail.MoneyAmount > 0 {
			right.WriteString(fmt.Sprintf("%d %s", m.Mail.MoneyAmount,
				style.HighlightStyle.Render(string(art.CharGrynars))))
		}

		right.WriteString("\n")

		if len(m.Attachments) > 0 {
			right.WriteString("\n")
			for i, a := range m.Attachments {
				if i > 0 {
					right.WriteString("\n")
				}

				right.WriteString(fmt.Sprintf("%c %dx %s",
					art.CharBullet, a.Amount, a.ItemName))
			}
		}

		if right.Len() > 0 && m.Mail.IsAgainstPayment {
			right.WriteString("\n")
			right.WriteString(style.DimBottomBorderStyle.
				Width(m.widthRight).Render(""))
		}

		if m.Mail.IsAgainstPayment {
			right.WriteString("\n\n")
			right.WriteString(style.TextStyle.Width(m.widthRight).
				Render(fmt.Sprintf(
					lang.L("The sender ask you to pay %d %c to access the attachments."),
					m.Mail.PaymentAmount, art.CharGrynars)))
		}
	}

	tui.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		style.FocusedStyle.Width(m.widthLeft).Render(left.String()),
		style.FocusedStyle.Width(m.widthRight).Render(right.String())))

	tui.WriteString("\n")

	if m.ErrMsg != nil {
		tui.WriteString(style.ErrorStyle.Render(m.ErrMsg.Error()))
	}

	tui.WriteString("\n")
	tui.WriteString(bubblehelp.View(style.LayoutWidth))

	return lipgloss.Place(
		context.ContentManager.ScreenWidth,
		context.ContentManager.ScreenHeight,
		lipgloss.Center, lipgloss.Center,
		tui.String(),
	)
}

func (m *Model) updateData() {
	req := request.MailGetOne(m.Mail.ID)

	resp, err := helper.SendRequest(req)

	if err != nil {
		m.ErrMsg = err
		return
	}

	mail := resp.Result().(*api.MailBasicResponse)

	m.Mail = mail
}

func (m *Model) updateAttachments() {
	if m.Mail == nil {
		return
	}

	req := request.MailGetAttachments(m.Mail.ID)

	resp, err := helper.SendRequest(req)

	if err != nil {
		m.ErrMsg = err
		return
	}

	m.Attachments = *resp.Result().(*[]api.MailAttachmentResponse)
}

func (m Model) updateKeymap() {
	if !m.Mail.IsAgainstPayment {
		bubblehelp.SetKeybindVisible(keybind.PKey, false)
	}

	if m.Mail.IsAgainstPayment || (len(m.Attachments) == 0 && m.Mail.MoneyAmount == 0) {
		bubblehelp.SetKeybindVisible(keybind.TKey, false)
	}

}

func (m *Model) payMail() {
	req := request.MailPay(m.Mail.ID)

	resp, err := helper.SendRequest(req)

	if err != nil {
		m.ErrMsg = err
		return
	}

	if resp.StatusCode() == http.StatusOK {
		m.updateData()
		m.updateKeymap()
	}
}

func (m *Model) transferAttachments() {
	req := request.MailTransferAttachments(m.Mail.ID)

	resp, err := helper.SendRequest(req)

	if err != nil {
		m.ErrMsg = err
		return
	}

	if resp.StatusCode() == http.StatusOK {
		m.updateData()
		m.updateAttachments()
		m.updateKeymap()
	}
}

func (m *Model) changeReadStatus(read bool) {
	req := request.MailSetRead(m.Mail.ID, read)

	resp, err := helper.SendRequest(req)

	if err != nil {
		m.ErrMsg = err
		return
	}

	if resp.StatusCode() == http.StatusOK {
		m.updateData()
		m.updateKeymap()
	}
}
