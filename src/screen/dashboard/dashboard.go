package dashboard

import (
	"farental/core/data"
	"farental/internal/context"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/screen"
	"farental/widget/characterinfo"
	"farental/widget/fullhelp"
	"farental/widget/help"
	"farental/widget/locationinfo"
	"farental/widget/runningtask"
	"farental/widget/simplelogviewer"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
)

const (
	tick time.Duration = 60
)

type Screen struct {
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

	t := orvyn.GetTheme()

	s.runningTask = runningtask.New()

	s.characterInfo = characterinfo.New()

	s.locationInfo = locationinfo.New()

	logStyle := simplelogviewer.Style{
		FocusedWidget: t.Style(theme.FocusedWidgetStyleID),
		BlurredWidget: t.Style(theme.BlurredWidgetStyleID),
		FocusedTitle:  t.Style(ftheme.TitleUnderlinedTextStyleID),
		BlurredTitle:  t.Style(ftheme.DimUnderlinedTextStyleID),
	}

	s.logEvent = simplelogviewer.New(lokyn.L("Events"))
	s.logEvent.Style = logStyle
	s.logEvent.OnBlur()

	s.logChat = simplelogviewer.New(lokyn.L("Chat"))
	s.logChat.Style = logStyle
	s.logChat.OnBlur()

	s.logCharacters = simplelogviewer.New(lokyn.L("Characters"))
	s.logCharacters.Style = logStyle
	s.logCharacters.OnBlur()

	s.help = help.New()

	s.fullHelp = fullhelp.New()
	s.fullHelp.SetActive(false)

	s.statusMessage = statusmessage.New()

	s.socialLayout = layout.NewHBoxGrowLayout(1, 0,
		s.logChat, s.logCharacters,
	)

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(
			35,
			t.Size(ftheme.LayoutWidthSizeID),
			10,
			s.runningTask,
			s.characterInfo,
			s.locationInfo,
			s.logEvent,
			layout.NewPileLayout(
				s.socialLayout,
				s.fullHelp,
			),
			s.statusMessage,
			s.help,
		),
	)

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.logEvent)
	s.focusManager.Add(s.logChat)
	s.focusManager.Add(s.logCharacters)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextGameDashboard)

	s.logEvent.SetTitle(lokyn.L("Events"))
	s.logChat.SetTitle(lokyn.L("Chat"))
	s.logCharacters.SetTitle(lokyn.L("Characters"))

	s.statusMessage.Reset()

	switch param := i.(type) {
	case error:
		s.statusMessage.SetError(param)
	case StatusMessageParam:
		s.statusMessage.SetMessage(param.Content, param.Type)
	}

	s.logEvent.SetContent(make([]string, 0))

	s.updateData()

	s.focusManager.Focus(0)

	s.showHelp(false)

	cmd := s.runningTask.Init()

	return tea.Batch(cmd, orvyn.TickCmd(tick, s.tickTag))
}

func (s *Screen) OnExit() any {
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

	s.fullHelp.SetTitle(lokyn.L("Location services"))

	// Activate keybind based on available features
	bubblehelp.SetKeybindVisible(keybind.BKey,
		context.CharacterInfo.Location.HaveFeature(string(data.FeatureBank)))
	bubblehelp.SetKeybindVisible(keybind.RKey,
		context.CharacterInfo.Location.HaveFeature(string(data.FeatureTavern)))
	bubblehelp.SetKeybindVisible(keybind.TKey,
		context.CharacterInfo.Location.HaveFeature(string(data.FeatureTavern)))
	bubblehelp.SetKeybindVisible(keybind.SKey,
		context.CharacterInfo.Location.HaveFeature(string(data.FeatureMerchant)))
	bubblehelp.SetKeybindVisible(keybind.MKey,
		context.CharacterInfo.Location.HaveFeature(string(data.FeatureMailbox)))
}

func (s *Screen) hideLocationService() {
	bubblehelp.SwitchContext(keybind.ContextGameDashboard)
	bubblehelp.ShowAll = false
	s.showHelp(false)

	s.fullHelp.SetTitle(lokyn.L("Help"))
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
		if s.checkRunningTask() {
			s.showLocationService()
		}

	case key.Matches(msg, keybind.BKeyCtrl):
		return orvyn.SwitchScreen(screen.IDSendFeedback), true

	case key.Matches(msg, keybind.TKey):
		if s.checkRunningTask() {
			return orvyn.SwitchScreen(screen.IDTravel), true
		}

	case key.Matches(msg, keybind.AKey):
		if s.checkRunningTask() {
			return orvyn.SwitchScreen(screen.IDActivity), true
		}

	case key.Matches(msg, keybind.FKey):
		if s.checkRunningTask() {
			return orvyn.SwitchScreen(screen.IDFight), true
		}

	case key.Matches(msg, keybind.CKey):
		if s.checkRunningTask() {
			return orvyn.SwitchScreen(screen.IDCraft), true
		}

	case key.Matches(msg, keybind.YKey):
		return orvyn.SwitchScreen(screen.IDChat), true

	case key.Matches(msg, keybind.IKey):
		return orvyn.SwitchScreen(screen.IDInventory), true

	case key.Matches(msg, keybind.HKey):
		return orvyn.SwitchScreen(screen.IDCharacterSheet), true

	case key.Matches(msg, keybind.NKey):
		if s.checkRunningTask() {
			return orvyn.SwitchScreen(screen.IDNpc), true
		}

	case key.Matches(msg, keybind.MKey):
		return orvyn.SwitchScreen(screen.IDLocationInfo), true

	case key.Matches(msg, keybind.SKey):
		if s.checkRunningTask() {
			return orvyn.SwitchScreen(screen.IDScriptExplorer), true
		}

	case key.Matches(msg, keybind.UKey):
		return orvyn.SwitchScreen(screen.IDUserSettings), true
	}

	return nil, false
}

func (s *Screen) servicesKeyHandler(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch {
	case key.Matches(msg, keybind.Esc):
		s.hideLocationService()

	case key.Matches(msg, keybind.SKey):
		if bubblehelp.IsKeybindVisible(keybind.SKey) {
			return orvyn.SwitchScreen(screen.IDShop), false
		}

	case key.Matches(msg, keybind.BKey):
		if bubblehelp.IsKeybindVisible(keybind.BKey) {
			return orvyn.SwitchScreen(screen.IDBank), false
		}

	case key.Matches(msg, keybind.TKey):
		if bubblehelp.IsKeybindVisible(keybind.TKey) {
			s.tavernSleep()

			return nil, true
		}

	case key.Matches(msg, keybind.RKey):
		if bubblehelp.IsKeybindVisible(keybind.RKey) {
			s.tavernRegen()

			return nil, true
		}

	case key.Matches(msg, keybind.MKey):
		if bubblehelp.IsKeybindVisible(keybind.MKey) {
			return orvyn.SwitchScreen(screen.IDMailBox), false
		}

	}

	return nil, false
}

func (s *Screen) checkRunningTask() bool {
	if context.RunningTask != nil {
		s.statusMessage.SetMessage(lokyn.L("A task is running. Claim the reward before doing this."), statusmessage.InformationMessage)
		return false
	}

	return true
}
