package gamedashboard

import (
	"errors"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/config"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/model"
	"farental/model/widget/charactervitalinfo"
	"farental/model/widget/locationinfo"
	"farental/model/widget/runningtask"
	"farental/model/widget/simplelogviewer"
	"farental/model/widget/widgetcontainer"
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-resty/resty/v2"
	"log"
	"strings"
	"time"
)

const (
	tick time.Duration = 15
)

type Model struct {
	Help          help.Model
	HelpContainer widgetcontainer.Model
	FullHelpStyle lipgloss.Style
	Keymap        config.ModularKeyMap
	ErrMsg        error

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

	m.FullHelpStyle = style.ContainerStyle.Width(style.LayoutWidth).
		Height(14)

	m.Help = help.New()
	m.Help.Styles.FullKey = style.TitleStyle
	m.Help.Styles.FullDesc = style.DimTextStyle

	m.HelpContainer = widgetcontainer.New(
		nil, lang.L("Help"), style.LayoutWidth, 14)

	m.Keymap = config.ModularKeyMap{}

	m.Keymap.SetBindings([][]key.Binding{
		{
			keybind.Travels,
			keybind.Activities,
			keybind.Crafts,
			keybind.Fights,
			keybind.LocationServices,
			keybind.Npcs,
		},
		{
			keybind.Scripts,
			keybind.Inventory,
			keybind.Claim,
			keybind.ChangeCharacter,
			keybind.Quit,
			keybind.HelpClose,
		},
	})

	m.Keymap.SetEssentialBindings([]key.Binding{
		keybind.Claim,
		keybind.Quit,
		keybind.HelpMore,
	})

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
		m.resetError()
		switch {
		case key.Matches(msg, keybind.HelpMore):
			m.Help.ShowAll = !m.Help.ShowAll
			return m, nil
		case key.Matches(msg, keybind.Claim):
			m.claim()
			return m, nil
		case key.Matches(msg, keybind.Quit):
			return m, tea.Quit
		case key.Matches(msg, keybind.Back):
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
		case key.Matches(msg, keybind.Fights):
			if context.RunningTask != nil {
				m.runningTaskError()
				return m, nil
			}

			return context.ContentManager.
				SwitchContent(m, model.ContentFightSelection)
		case key.Matches(msg, keybind.Crafts):
			if context.RunningTask != nil {
				m.runningTaskError()
				return m, nil
			}

			return context.ContentManager.
				SwitchContent(m, model.ContentCraftSelection)
		case key.Matches(msg, keybind.Chat):
			return context.ContentManager.
				SwitchContent(m, model.ContentChat)

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

	if !m.Help.ShowAll {
		bottom.WriteString(lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.JoinHorizontal(lipgloss.Center,
				m.ChatViewerContainer.View(),
				m.CharactersVisibleContainer.View()),
			m.Help.View(m.Keymap)))
	} else {
		bottom.WriteString(m.HelpContainer.ViewContent(
			m.Help.View(m.Keymap),
			lipgloss.Center, lipgloss.Center))
		bottom.WriteString("\n")
	}

	error := ""

	if m.ErrMsg != nil {
		error = fmt.Sprintf("%v\n", m.ErrMsg.Error())
	}

	tui = lipgloss.JoinVertical(lipgloss.Center,
		style.ErrorStyle.Render(error),
		style.ContainerStyle.Render(m.RunningTask.View()),
		style.ContainerStyle.Render(m.CharacterVitalInfo.View()),
		style.ContainerStyle.Render(m.LocationInfo.View()),
		m.EventLogViewerContainer.View(),
		bottom.String())

	return lipgloss.Place(
		context.ContentManager.ScreenWidth,
		context.ContentManager.ScreenHeight,
		lipgloss.Center,
		lipgloss.Center,
		tui)
}

func (m *Model) UpdateData() {
	var req *resty.Request

	req = request.CharacterGetInfo()

	resp, err := req.Send()

	if err != nil {
		m.ErrMsg = helper.ConnectionError()
		return
	}

	m.ErrMsg = helper.ExtractError(resp)

	if m.ErrMsg != nil {
		return
	}

	characterInfo := resp.Result().(*api.CharacterInfoResponse)

	// If the character changed of location
	if context.CharacterInfo == nil ||
		context.CharacterInfo.Location.ID != characterInfo.Location.ID {
		context.ChatContent = make([]string, 0)
	}

	context.CharacterID = characterInfo.ID
	context.CharacterInfo = characterInfo

	req = request.CharacterGetCurrencyAmount(api.Grynars)

	resp, err = req.Send()

	if err != nil {
		m.ErrMsg = helper.ConnectionError()
		return
	}

	m.ErrMsg = helper.ExtractError(resp)

	if m.ErrMsg != nil {
		return
	}

	currencyResp := resp.Result().(*api.CurrencyResponse)

	m.CharacterVitalInfo.UpdateData(characterInfo, currencyResp.Amount)
	m.LocationInfo.UpdateData(&characterInfo.Location)
	m.updateEventLog()
	m.updateChat()
	m.updateCharactersConnected()
	m.updateRunningTask()
}

func (m *Model) updateEventLog() {
	var req *resty.Request
	var queryParam string

	req = request.CharacterGetEventLog()

	queryParam = ""
	length := len(m.EventLogViewer.Content)

	if length > 0 {
		queryParam = m.lastEventLogTimestamp.Format(time.DateTime)
	}

	req.SetQueryParam("lastTimestamp", queryParam)

	resp, err := req.Send()

	if err != nil {
		log.Println(err)
		return
	}

	eventLog := resp.Result().(*api.EventLogResponse)

	if len(eventLog.Entries) == 0 {
		return
	}

	for _, entry := range eventLog.Entries {
		m.EventLogViewer.AddContent(fmt.Sprintf("%s - %s",
			style.TitleStyle.Render(
				entry.Timestamp.Format("01.02.2006 15:04:05")),
			entry.Value))
	}

	m.EventLogViewerContainer.UpdateContent(m.EventLogViewer)

	m.lastEventLogTimestamp = eventLog.Entries[len(eventLog.Entries)-1].Timestamp
}

func (m *Model) updateChat() {
	context.UpdateChat()

	if len(context.ChatContent) > len(m.ChatViewer.Content) {
		m.ChatViewer.SetContent(context.ChatContent)
		m.ChatViewerContainer.UpdateContent(m.ChatViewer)
	}
}

func (m *Model) updateCharactersConnected() {
	var req *resty.Request
	var str []string

	req = request.LocationGetCharacters()

	resp, err := req.Send()

	if err != nil {
		log.Println(err)
		return
	}

	characters := *resp.Result().(*[]api.CharacterBasicWithActivityResponse)

	str = make([]string, 0)

	for _, character := range characters {
		str = append(str,
			style.TitleStyle.Render(fmt.Sprintf("%s %s - %s\n  %s",
				character.FirstName, character.LastName,
				character.RaceName, character.CurrentActivityTitle)))
	}

	m.CharactersVisible.SetContent(str)

	m.CharactersVisibleContainer.UpdateContent(m.CharactersVisible)
}

func (m *Model) updateRunningTask() {
	var req *resty.Request

	req = request.TaskGetRunning()

	resp, err := req.Send()

	if err != nil {
		log.Println(err)
		return
	}

	if resp.StatusCode() == 404 {
		context.RunningTask = nil
		return
	}

	task := resp.Result().(*api.TaskResponse)

	context.RunningTask = task
}

func (m *Model) runningTaskError() {
	if context.RunningTask.IsRunning {
		m.ErrMsg = errors.New(lang.L("A task is currently running."))
	} else {
		m.ErrMsg = errors.New(lang.L("Please claim your reward first."))
	}
}

func (m *Model) resetError() {
	m.ErrMsg = nil
}

func (m *Model) claim() {
	if context.RunningTask == nil {
		return
	}

	if context.RunningTask.IsRunning {
		m.runningTaskError()
	}

	req := request.TaskClaim()

	resp, err := req.Send()

	if err != nil {
		log.Println(err)
		return
	}

	if resp.StatusCode() != 200 {
		log.Println(resp.StatusCode(), resp.Error())
		return
	}

	context.RunningTask = nil
	m.UpdateData()
}
