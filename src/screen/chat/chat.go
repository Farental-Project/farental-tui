package chat

import (
	"errors"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/internal/ticker"
	"farental/widget/help"
	"farental/widget/simplelogviewer"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/textarea"
)

type Screen struct {
	ticker *ticker.Ticker

	title *orvyn.SimpleRenderable

	logChat *simplelogviewer.Widget

	input *textarea.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	logStyle := simplelogviewer.Style{
		FocusedWidget: t.Style(theme.FocusedWidgetStyleID),
		BlurredWidget: t.Style(theme.BlurredWidgetStyleID),
		FocusedTitle:  t.Style(ftheme.TitleUnderlinedTextStyleID),
		BlurredTitle:  t.Style(ftheme.DimUnderlinedTextStyleID),
	}

	s.title = orvyn.NewSimpleRenderable("Chat")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.logChat = simplelogviewer.New("")
	s.logChat.Style = logStyle
	s.logChat.Keybind.ScrollUp = keybind.PrevPage
	s.logChat.Keybind.ScrollDown = keybind.NextPage
	s.logChat.OnBlur()

	s.input = textarea.New()
	s.input.ShowLineNumbers = false
	s.input.SetMinSize(orvyn.NewSize(10, 5))
	s.input.KeyMap.InsertNewline = keybind.YKeyCtrl
	s.input.Focus()

	s.statusMessage = statusmessage.New()

	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(10, 4),
			2,
			s.title,
			orvyn.VGap,
			s.logChat,
			s.input,
			s.statusMessage,
			s.help,
		),
	)

	s.ticker = ticker.New(15, s.loadChat)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextChat)

	s.title.SetValue(lokyn.L("Chat"))

	cmd := s.input.Init()

	s.loadChat()

	return tea.Batch(s.ticker.Start(), cmd)
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

		case key.Matches(msg, keybind.Enter):
			s.sendMessage()
			return nil

		}

	case orvyn.TickMsg:
		handled, cmd := s.ticker.Handle(msg)

		if !handled {
			return nil
		}

		return cmd

	}

	cmd := s.input.Update(msg)

	s.logChat.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) sendMessage() {
	var message string

	message = s.input.Value()

	if len(message) == 0 {
		s.statusMessage.SetError(
			errors.New(lokyn.L("Can't send empty messages")))
		return
	}

	message = helper.RemoveEmptyLines(message, 5)

	req := request.ChatSendMessage()

	req.Body = api.ChatMessageBody{
		Message: message,
	}

	_, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.input.SetValue("")
	s.loadChat()
}

func (s *Screen) loadChat() {
	context.UpdateChat()

	ctxContent := context.ChatContent
	content := s.logChat.GetContent()

	if len(ctxContent) == 0 {
		s.logChat.SetContent(ctxContent)
		return
	}

	if len(content) == 0 {
		s.logChat.SetContent(ctxContent)
		return
	}

	if ctxContent[len(ctxContent)-1] != content[len(content)-1] {
		s.logChat.SetContent(ctxContent)
	}
}
