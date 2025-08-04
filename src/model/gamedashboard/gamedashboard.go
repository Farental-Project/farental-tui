package gamedashboard

import (
	"errors"
	"farental/core/data"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/model"
	"farental/model/widgetmodel/charactervitalinfo"
	"farental/model/widgetmodel/locationinfo"
	"farental/model/widgetmodel/runningtask"
	"farental/model/widgetmodel/simplelogviewer"
	"farental/model/widgetmodel/widgetcontainer"
	"farental/style"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"log"
	"strings"
	"time"
)

const (
	tick time.Duration = 15
)

type Model struct {
	HelpContainer widgetcontainer.Model
	ErrMsg        error
	SuccessMsg    string

	tickTag uint

	RunningTask        runningtask.Model
	CharacterVitalInfo charactervitalinfo.Model
	LocationInfo       locationinfo.Model

	EventLogViewer          simplelogviewer.Model
	EventLogViewerContainer widgetcontainer.Model

	ChatViewer          simplelogviewer.Model
	ChatViewerContainer widgetcontainer.Model

	CharactersVisible          simplelogviewer.Model
	CharactersVisibleContainer widgetcontainer.Model

	lastEventLogTimestamp time.Time
}

func New() Model {
	m := Model{
		RunningTask:        runningtask.New(style.LayoutWidth),
		CharacterVitalInfo: charactervitalinfo.New(style.LayoutWidth),
		LocationInfo:       locationinfo.New(style.LayoutWidth),
		EventLogViewer:     simplelogviewer.New(style.LayoutWidth, 12),
		ChatViewer:         simplelogviewer.New(48, 12),
		CharactersVisible:  simplelogviewer.New(25, 12),
	}

	m.EventLogViewerContainer = widgetcontainer.New(
		m.EventLogViewer,
		lang.L("Event log"), style.LayoutWidth, 14)
	m.ChatViewerContainer = widgetcontainer.New(
		m.ChatViewer,
		lang.L("Chat"), 48, 14)
	m.CharactersVisibleContainer = widgetcontainer.New(
		m.CharactersVisible,
		lang.L("Characters in location"), 25, 14)

	m.HelpContainer = widgetcontainer.New(
		nil, lang.L("Help"), style.LayoutWidth, 14)

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(model.InitCmd,
		model.TickCmd(m, tick, m.tickTag))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var mod tea.Model

	defer context.ContentManager.UpdateCurrentContent(m)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.resetMsg()

		switch bubblehelp.CurrentContext {
		case model.ContextGameDashboard:
			mod, mes := m.gameKeyHandler(msg)

			if mod != nil {
				return mod, mes
			}

		case model.ContextLocationServices:
			mod, mes := m.servicesKeyHandler(msg)

			if mod != nil {
				return mod, mes
			}
		}

	case model.TickMsg:
		if msg.Tag != m.tickTag {
			return m, nil
		}

		m.UpdateData()

		m.tickTag++
		return m, model.TickCmd(m, tick, m.tickTag)
	case model.InitMsg:
		m.UpdateData()

		cmd = m.RunningTask.Init()

		bubblehelp.SwitchContext(model.ContextGameDashboard)

		return m, cmd
	}

	context.ContentManager.Update(msg)

	// Spinner need update
	mod, cmd = m.RunningTask.Update(msg)
	m.RunningTask = mod.(runningtask.Model)

	return m, cmd
}

func (m Model) View() string {
	var tui string
	var bottom strings.Builder

	if !bubblehelp.ShowAll {
		bottom.WriteString(lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.JoinHorizontal(lipgloss.Center,
				m.ChatViewerContainer.View(),
				m.CharactersVisibleContainer.View())))

		m.renderMsg(&bottom)

		bottom.WriteString("\n")
		bottom.WriteString(bubblehelp.View(style.LayoutWidth))
	} else {
		bottom.WriteString(m.HelpContainer.ViewContent(
			bubblehelp.ViewAll(
				bubblehelp.GetCurrentContextKeymap(),
				style.LayoutWidth),
			lipgloss.Center, lipgloss.Center))

		m.renderMsg(&bottom)

		bottom.WriteString("\n")
	}

	tui = lipgloss.JoinVertical(lipgloss.Center,
		style.FocusedStyle.Render(m.RunningTask.View()),
		style.FocusedStyle.Render(m.CharacterVitalInfo.View()),
		style.FocusedStyle.Render(m.LocationInfo.View()),
		m.EventLogViewerContainer.View(),
		bottom.String())

	return lipgloss.Place(
		context.ContentManager.ScreenWidth,
		context.ContentManager.ScreenHeight,
		lipgloss.Center,
		lipgloss.Center,
		tui)
}

func (m *Model) runningTaskError() {
	if context.RunningTask.IsRunning {
		m.ErrMsg = errors.New(lang.L("A task is currently running."))
	} else {
		m.ErrMsg = errors.New(lang.L("Please claim your reward first."))
	}
}

func (m *Model) resetMsg() {
	m.ErrMsg = nil
	m.SuccessMsg = ""
}

func (m *Model) renderMsg(b *strings.Builder) {
	b.WriteString("\n")
	if m.ErrMsg != nil || len(m.SuccessMsg) > 0 {
		switch {
		case m.ErrMsg != nil:
			b.WriteString(style.ErrorStyle.Render(m.ErrMsg.Error()))
		case len(m.SuccessMsg) > 0:
			b.WriteString(style.TitleStyle.Render(m.SuccessMsg))
		}
	}
}

func (m *Model) claim() {
	if context.RunningTask == nil {
		return
	}

	if context.RunningTask.IsRunning {
		m.runningTaskError()
	}

	_, err := helper.SendRequest(request.TaskClaim())

	if err != nil {
		log.Println(err)
		return
	}

	context.RunningTask = nil
	m.UpdateData()
}

func (m *Model) showLocationService() {
	bubblehelp.SwitchContext(model.ContextLocationServices)
	bubblehelp.ShowAll = true

	m.HelpContainer.Title = lang.L("Location services")

	// Activate keybind based on available features
	bubblehelp.SetKeybindVisible(keybind.RKey,
		context.CharacterInfo.Location.HaveFeature(string(data.FeatureTavern)))
	bubblehelp.SetKeybindVisible(keybind.MKey,
		context.CharacterInfo.Location.HaveFeature(string(data.FeatureMailbox)))
}

func (m *Model) hideLocationService() {
	bubblehelp.SwitchContext(model.ContextGameDashboard)
	bubblehelp.ShowAll = false

	m.HelpContainer.Title = lang.L("Help")
}

func (m *Model) gameKeyHandler(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keybind.Help):
		bubblehelp.ShowAll = !bubblehelp.ShowAll
		return m, nil

	case key.Matches(msg, keybind.Space):
		m.claim()
		return m, nil

	case key.Matches(msg, keybind.Quit):
		return m, tea.Quit

	case key.Matches(msg, keybind.Esc):
		return context.ContentManager.
			SwitchContent(m, model.ContentCharacterSelection)

	case key.Matches(msg, keybind.Travels):
		if context.RunningTask != nil {
			m.runningTaskError()
			return m, nil
		}

		return context.ContentManager.
			SwitchContent(m, model.ContentTravelSelection)

	case key.Matches(msg, keybind.Activities):
		if context.RunningTask != nil {
			m.runningTaskError()
			return m, nil
		}

		return context.ContentManager.
			SwitchContent(m, model.ContentActivitySelection)

	case key.Matches(msg, keybind.Inventory):
		return context.ContentManager.
			SwitchContent(m, model.ContentInventory)

	case key.Matches(msg, keybind.FKey):
		if context.RunningTask != nil {
			m.runningTaskError()
			return m, nil
		}

		return context.ContentManager.
			SwitchContent(m, model.ContentFightSelection)

	case key.Matches(msg, keybind.CKey):
		if context.RunningTask != nil {
			m.runningTaskError()
			return m, nil
		}

		return context.ContentManager.
			SwitchContent(m, model.ContentCraftSelection)

	case key.Matches(msg, keybind.YKey):
		return context.ContentManager.
			SwitchContent(m, model.ContentChat)

	case key.Matches(msg, keybind.CharacterSheet):
		return context.ContentManager.
			SwitchContent(m, model.ContentCharacterSheet)

	case key.Matches(msg, keybind.LocationServices):
		m.showLocationService()

		return m, nil
	}

	return nil, nil
}

func (m *Model) servicesKeyHandler(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keybind.Quit):
		return m, tea.Quit

	case key.Matches(msg, keybind.Esc):
		m.hideLocationService()

		return m, nil

	case key.Matches(msg, keybind.RKey):
		if bubblehelp.IsKeybindVisible(keybind.RKey) {
			m.tavernSleep()
			return m, nil
		}

	case key.Matches(msg, keybind.MKey):
		if bubblehelp.IsKeybindVisible(keybind.MKey) {
			return context.ContentManager.SwitchContent(
				m, model.ContentMailbox)
		}
	}

	return nil, nil
}
