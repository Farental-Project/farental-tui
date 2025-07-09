package mailbox

import (
	"farental/core/data/api"
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
	"io"
	"strings"
)

type ListItem struct {
	Mail api.MailBasicResponse
}

func NewListItem(mail api.MailBasicResponse) ListItem {
	return ListItem{mail}
}

func (i ListItem) FilterValue() string {
	var b strings.Builder

	b.WriteString(i.Mail.DeliveredAt.Format(viper.GetString("datetimeformat")))
	b.WriteString(" ")
	b.WriteString(i.Mail.SenderName)
	b.WriteString(" ")
	b.WriteString(i.Mail.Subject)
	b.WriteString(" ")
	b.WriteString(i.Mail.Content)

	return b.String()
}

type ListItemDelegate struct{}

func (l ListItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(ListItem)

	if !ok {
		return
	}

	var s lipgloss.Style
	var left strings.Builder
	var right strings.Builder
	var width int

	width = m.Width() - 2

	if index == m.Index() {
		s = style.FocusedStyle
	} else {
		s = style.BlurredStyle
	}

	left.WriteString(style.NormalStyle.Render(fmt.Sprint(i.Mail.DeliveredAt.Format(viper.GetString("datetimeformat")))))
	left.WriteString("\n")
	left.WriteString(style.TitleStyle.Render(i.Mail.SenderName))
	left.WriteString("\n")
	left.WriteString(style.DimTextStyle.Render(i.Mail.Subject))

	if !i.Mail.IsRead {
		right.WriteString(style.HighlightStyle.Render("•"))
	} else {
		right.WriteString("")
	}

	tui := s.Width(m.Width() - 2).Height(l.Height()).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			style.TextStyle.Width(width-2).
				AlignHorizontal(lipgloss.Left).
				Render(left.String()),
			style.TextStyle.Width(1).
				AlignHorizontal(lipgloss.Right).
				Render(right.String())))

	fmt.Fprint(w, tui)
}

func (l ListItemDelegate) Height() int {
	return 3
}

func (l ListItemDelegate) Spacing() int {
	return 0
}

func (l ListItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
