package dashboard

import (
	"farental/core/data"
	"farental/internal/context"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	layout "farental/layout"
	"farental/screen"
	"farental/style"
	"farental/widget/characterinfo"
	"farental/widget/fullhelp"
	"farental/widget/help"
	"farental/widget/locationinfo"
	"farental/widget/runningtask"
	"farental/widget/simplelogviewer"
	"farental/widget/statusmessage"
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

	runningTask *runningtask.Widget

	characterInfo *characterinfo.Widget

	locationInfo *locationinfo.Widget

	logEvent *simplelogviewer.Widget

	logChat *simplelogviewer.Widget

	logCharacters *simplelogviewer.Widget

	help *help.Widget

	fullHelp *fullhelp.Widget

	statusMessage *statusmessage.Widget

	lastEventLogTimestamp time.Time

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout

	socialLayout *layout.HBoxGrowLayout
}

func New() *Screen {
	s := new(Screen)

	s.runningTask = runningtask.New()

	s.characterInfo = characterinfo.New()

	s.locationInfo = locationinfo.New()

	logStyle := simplelogviewer.Style{
		FocusedWidget: style.FocusedStyle,
		BlurredWidget: style.BlurredStyle,
		FocusedTitle:  style.HighlightUnderlinedTitleStyle,
		BlurredTitle:  style.DimUnderlinedTitleStyle,
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

	s.fullHelp = fullhelp.New()
	s.fullHelp.SetActive(false)

	s.statusMessage = statusmessage.New()

	s.socialLayout = layout.NewHBoxGrowLayout(1, 0,
		[]orvyn.Renderable{
			s.logChat, s.logCharacters,
		})

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(
			35,
			style.LayoutWidth,
			10,
			[]orvyn.Renderable{
				s.runningTask,
				s.characterInfo,
				s.locationInfo,
				s.logEvent,
				layout.NewPileLayout([]orvyn.Renderable{
					s.socialLayout,
					s.fullHelp,
				}),
				s.statusMessage,
				s.help,
			}),
	)

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.logEvent)
	s.focusManager.Add(s.logChat)
	s.focusManager.Add(s.logCharacters)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextGameDashboard)

	s.updateData()

	s.focusManager.Focus(0)

	s.showHelp(false)

	cmd := s.runningTask.Init()

	return tea.Batch(cmd, orvyn.TickCmd(tick, s.tickTag))
}

func (s *Screen) OnExit() interface{} {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s.statusMessage.Reset()

		if key.Matches(msg, keybind.Quit) {
			return tea.Quit
		}

		switch bubblehelp.CurrentContext {
		case keybind.ContextGameDashboard:
			c, ok := s.gameKeyHandler(msg)

			if ok {
				return c
			}
		case keybind.ContextLocationServices:
			c, ok := s.servicesKeyHandler(msg)

			if ok {
				s.hideLocationService()
				return c
			}
		}

	case orvyn.TickMsg:
		if msg.Tag != s.tickTag {
			return nil
		}

		s.updateData()

		s.tickTag++
		return orvyn.TickCmd(tick, s.tickTag)
	}

	s.focusManager.Update(msg)

	cmd := s.runningTask.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) showHelp(b bool) {
	s.help.SetActive(!b)
	s.socialLayout.SetActive(!b)
	s.fullHelp.SetActive(b)
}

func (s *Screen) showLocationService() {
	bubblehelp.SwitchContext(keybind.ContextLocationServices)
	bubblehelp.ShowAll = true
	s.showHelp(true)

	s.fullHelp.SetTitle(lang.L("Location services"))

	// Activate keybind based on available features
	bubblehelp.SetKeybindVisible(keybind.RKey,
		context.CharacterInfo.Location.HaveFeature(string(data.FeatureTavern)))
	bubblehelp.SetKeybindVisible(keybind.MKey,
		context.CharacterInfo.Location.HaveFeature(string(data.FeatureMailbox)))
}

func (s *Screen) hideLocationService() {
	bubblehelp.SwitchContext(keybind.ContextGameDashboard)
	bubblehelp.ShowAll = false
	s.showHelp(false)

	s.fullHelp.SetTitle(lang.L("Help"))
}

func (s *Screen) gameKeyHandler(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch {
	case key.Matches(msg, keybind.Esc):
		return orvyn.SwitchScreen(screen.IDCharacterSelection), true

	case key.Matches(msg, keybind.Help):
		bubblehelp.ShowAll = !bubblehelp.ShowAll
		s.showHelp(bubblehelp.ShowAll)

		return nil, true

	case key.Matches(msg, keybind.Space):
		s.claim()

		return nil, true

	case key.Matches(msg, keybind.LKey):
		s.showLocationService()

	case key.Matches(msg, keybind.TKey):
		return orvyn.SwitchScreen(screen.IDTravel), true

	case key.Matches(msg, keybind.AKey):
		return orvyn.SwitchScreen(screen.IDActivity), true

	case key.Matches(msg, keybind.FKey):
		return orvyn.SwitchScreen(screen.IDFight), true

	case key.Matches(msg, keybind.CKey):
		return orvyn.SwitchScreen(screen.IDCraft), true

	case key.Matches(msg, keybind.YKey):
		return orvyn.SwitchScreen(screen.IDChat), true

	case key.Matches(msg, keybind.IKey):
		return orvyn.SwitchScreen(screen.IDInventory), true

	case key.Matches(msg, keybind.HKey):
		return orvyn.SwitchScreen(screen.IDCharacterSheet), true

	}

	return nil, false
}

func (s *Screen) servicesKeyHandler(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch {
	case key.Matches(msg, keybind.Esc):
		s.hideLocationService()

	case key.Matches(msg, keybind.RKey):
		if bubblehelp.IsKeybindVisible(keybind.RKey) {
			s.tavernSleep()

			return nil, true
		}

	case key.Matches(msg, keybind.MKey):
		if bubblehelp.IsKeybindVisible(keybind.MKey) {
			return orvyn.SwitchScreen(screen.IDMailBox), true
		}

	}

	return nil, false
}
