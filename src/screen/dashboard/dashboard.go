package dashboard

import (
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/internal/orvyn/layout"
	"farental/style"
	"farental/widget/help"
	"farental/widget/runningtask"
	"farental/widget/simplelogviewer"
	"farental/widget/statusmessage"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

type Screen struct {
	orvyn.BaseScreen

	runningTask *runningtask.Widget

	logEvent *simplelogviewer.Widget

	logChat *simplelogviewer.Widget

	logCharacters *simplelogviewer.Widget

	help *help.Widget

	statusMessage *statusmessage.Widget

	lastEventLogTimestamp time.Time

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	s.runningTask = runningtask.New()

	logStyle := simplelogviewer.Style{
		FocusedWidget: style.FocusedStyle,
		BlurredWidget: style.BlurredStyle,
		FocusedTitle:  style.TitleStyle,
		BlurredTitle:  style.DimBottomBorderStyle,
	}

	s.logEvent = simplelogviewer.New(lang.L("Events"))
	s.logEvent.Style = logStyle
	s.logEvent.OnBlur()

	s.logChat = simplelogviewer.New(lang.L("Chat"))
	s.logChat.Style = logStyle
	s.logChat.OnBlur()

	s.logCharacters = simplelogviewer.New(lang.L("Characters"))
	s.logCharacters.Style = logStyle
	s.logCharacters.OnBlur()

	s.help = help.New()

	s.statusMessage = statusmessage.New()

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(
			35,
			style.LayoutWidth,
			10,
			[]orvyn.Renderable{
				s.runningTask,
				s.logEvent,
				s.logChat,
				s.logCharacters,
				s.statusMessage,
				s.help,
			}),
	)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
	return nil
}

func (s *Screen) OnExit() interface{} {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return tea.Quit
		}
	}

	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}
