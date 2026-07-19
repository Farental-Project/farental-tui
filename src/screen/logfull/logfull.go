package logfull

import (
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/internal/ticker"
	"farental/widget/help"
	"farental/widget/runningtask"
	"farental/widget/simplelogviewer"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
)

type Screen struct {
	logViewer *simplelogviewer.Widget

	runningTask *runningtask.Widget

	ticker *ticker.Ticker

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

	s.runningTask = runningtask.New()

	s.logViewer = simplelogviewer.New("log")
	s.logViewer.Style = logStyle
	s.logViewer.OnFocus()

	s.help = help.New()

	s.ticker = ticker.New(60, func() {
		s.runningTask.RefreshCurrentCharacter()
	})

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(10, 4), 1,
			s.runningTask,
			s.logViewer,
			s.help,
		),
	)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	data, ok := i.([]string)

	if !ok {
		return orvyn.SwitchToPreviousScreen()
	}

	s.logViewer.SetContent(data)
	s.logViewer.SetTitle(lokyn.L("Events"))

	s.runningTask.RefreshCurrentCharacter()

	return tea.Batch(s.runningTask.Init(), s.ticker.Start())
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Esc):
			return orvyn.SwitchToPreviousScreen()
		}

	case orvyn.TickMsg:
		handled, cmd := s.ticker.Handle(msg)

		if !handled {
			return nil
		}

		return cmd
	}

	s.logViewer.Update(msg)

	return s.runningTask.Update(msg)
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}
