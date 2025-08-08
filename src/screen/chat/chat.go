package chat

import (
	"errors"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/internal/orvyn/layout"
	"farental/style"
	"farental/widget/help"
	"farental/widget/simplelogviewer"
	"farental/widget/statusmessage"
	"farental/widget/textarea"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"time"
)

const (
	tick time.Duration = 15
)

type Screen struct {
	orvyn.BaseScreen

	tickTag uint

	title *orvyn.SimpleRenderable

	logChat *simplelogviewer.Widget

	input *textarea.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	logStyle := simplelogviewer.Style{
		FocusedWidget: style.FocusedStyle,
		BlurredWidget: style.BlurredStyle,
		FocusedTitle:  style.HighlightUnderlinedTitleStyle,
		BlurredTitle:  style.DimUnderlinedTitleStyle,
	}

	s.title = orvyn.NewSimpleRenderable(
		style.TitleStyle.Render(lang.L("Chat")),
	)

	s.logChat = simplelogviewer.New(lang.L("Chat"))
	s.logChat.Style = logStyle
	s.logChat.Keybind.ScrollUp = keybind.PrevPage
	s.logChat.Keybind.ScrollDown = keybind.NextPage
	s.logChat.OnBlur()

	s.input = textarea.New()
	s.input.ShowLineNumbers = false
	s.input.MinHeight = 3
	s.input.KeyMap.InsertNewline = keybind.YKeyCtrl
	s.input.Focus()

	s.statusMessage = statusmessage.New()

	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(10, 4),
			2,
			[]orvyn.Renderable{
				s.title,
				orvyn.VGap,
				s.logChat,
				s.input,
				s.statusMessage,
				s.help,
			},
		),
	)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextChat)

	cmd := s.input.Init()

	s.loadChat()

	return tea.Batch(orvyn.TickCmd(tick, s.tickTag), cmd)
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

		case key.Matches(msg, keybind.Enter):
			s.sendMessage()
			return nil

		}

	case orvyn.TickMsg:
		if msg.Tag != s.tickTag {
			return nil
		}

		s.loadChat()

		s.tickTag++
		return orvyn.TickCmd(tick, s.tickTag)

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
			errors.New(lang.L("Can't send empty messages")))
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
