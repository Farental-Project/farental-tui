package mailreader

import (
	"farental/art"
	"farental/core/data/api"
	"farental/internal/context"
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
	"strings"
)

type Model struct {
	Mail        *api.MailBasicResponse
	Attachments []api.MailAttachment

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
		content := context.ContentManager.GetContent(model.ContentMailbox)

		mailbox := content.(mailbox.Model)

		m.Mail = mailbox.SelectedMail

		m.VPContent.SetContent(m.Mail.Content)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return m, tea.Quit
		case key.Matches(msg, keybind.Esc):
			return context.ContentManager.Back(m)
		}
	}

	context.ContentManager.Update(msg)

	return m, nil
}

func (m Model) View() string {
	var left, right strings.Builder

	if m.Mail == nil {
		return ""
	}

	left.WriteString(fmt.Sprintf("%s : %s",
		style.TitleStyle.Render(lang.L("From")),
		m.Mail.SenderName))
	left.WriteString("\n")
	left.WriteString(style.DimBottomBorderStyle.
		Width(m.widthLeft).
		Render(fmt.Sprintf("%s : %s",
			style.TitleStyle.Render(lang.L("Subject")),
			m.Mail.Subject)))
	left.WriteString("\n")
	left.WriteString(m.VPContent.View())

	if !m.Mail.HaveAttachments && m.Mail.MoneyAmount == 0 {
		right.WriteString(style.TextStyle.Width(m.widthRight).
			Render("No attachments"))
	} else {
		if m.Mail.MoneyAmount > 0 {
			right.WriteString(fmt.Sprintf("%d %s", m.Mail.MoneyAmount,
				style.HighlightStyle.Render(string(art.CharGrynars))))
		}

		right.WriteString("\n")

		if right.Len() > 0 {
			right.WriteString(style.DimBottomBorderStyle.
				Width(m.widthRight).Render("\n"))
		}
	}

	tui := lipgloss.JoinHorizontal(lipgloss.Top,
		style.ContainerStyle.Render(left.String()),
		style.ContainerStyle.Render(right.String()))

	return lipgloss.Place(
		context.ContentManager.ScreenWidth,
		context.ContentManager.ScreenHeight,
		lipgloss.Center, lipgloss.Center,
		tui,
	)
}
