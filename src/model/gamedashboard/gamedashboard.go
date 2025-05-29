package gamedashboard

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/config"
	"farental/internal/context"
	"farental/internal/lang"
	"farental/model"
	"farental/model/widget/charactervitalinfo"
	"farental/model/widget/locationinfo"
	"farental/model/widget/runningtask"
	"farental/model/widget/simplelogviewer"
	"farental/style"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-resty/resty/v2"
	"log"
	"time"
)

type Model struct {
	RunningTask        runningtask.Model
	CharacterVitalInfo charactervitalinfo.Model
	LocationInfo       locationinfo.Model
	EventLogViewer     simplelogviewer.Model
	ChatViewer         simplelogviewer.Model
	CharactersVisible  simplelogviewer.Model
	Keymap             config.ModularKeyMap

	lastEventLogTimestamp time.Time
	lastChatTimestamp     time.Time
}

func New() Model {
	m := Model{
		RunningTask:        runningtask.New(75),
		CharacterVitalInfo: charactervitalinfo.New(75),
		LocationInfo:       locationinfo.New(75),
		EventLogViewer: simplelogviewer.New(
			lang.L("Event log"), 75, 12),
		ChatViewer: simplelogviewer.New(
			lang.L("Chat"), 48, 12),
		CharactersVisible: simplelogviewer.New(
			lang.L("Characters in location"), 25, 12),
	}

	m.Keymap = config.ModularKeyMap{}

	m.Keymap.SetBindings([][]key.Binding{
		{},
		{
			config.Back,
			config.Help,
			config.Quit,
		},
	})

	return m
}

type tickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(15*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(model.InitCmd, doTick())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, config.Quit):
			return m, tea.Quit
		case key.Matches(msg, config.Back):
			return context.ContentManager.SwitchContent(
				model.ContentCharacterSelection)
		}
	case tickMsg:
		m.UpdateData()

		return m, doTick()
	case model.InitMsg:
		m.UpdateData()

		return m, nil
	}

	context.ContentManager.Update(msg)

	return m, nil
}

func (m Model) View() string {
	tui := lipgloss.JoinVertical(lipgloss.Center,
		style.ContainerStyle.Render(m.CharacterVitalInfo.View()),
		style.ContainerStyle.Render(m.LocationInfo.View()),
		style.ContainerStyle.Render(m.EventLogViewer.View()),
		lipgloss.JoinHorizontal(lipgloss.Center,
			style.ContainerStyle.Render(m.ChatViewer.View()),
			style.ContainerStyle.Render(m.CharactersVisible.View())))

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
		log.Println(err)
		return
	}

	characterInfo := resp.Result().(*api.CharacterInfoResponse)

	m.CharacterVitalInfo.UpdateData(characterInfo)
	m.LocationInfo.UpdateData(&characterInfo.Location)
	m.updateEventLog()
	m.updateChat()
	m.updateCharactersConnected()
	m.updateCharactersConnected()
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

	m.lastEventLogTimestamp = eventLog.Entries[len(eventLog.Entries)-1].Timestamp
}

func (m *Model) updateChat() {
	var req *resty.Request
	var queryParam string

	req = request.ChatGetMessages()

	queryParam = ""
	length := len(m.ChatViewer.Content)

	if length > 0 {
		queryParam = m.lastChatTimestamp.Format(time.DateTime)
	}

	req.SetQueryParam("lastTimestamp", queryParam)

	resp, err := req.Send()

	if err != nil {
		log.Println(err)
		return
	}

	chatMessages := *resp.Result().(*[]api.ChatMessageResponse)

	if len(chatMessages) == 0 {
		return
	}

	for _, message := range chatMessages {
		m.ChatViewer.AddContent(fmt.Sprintf("%s %s - %s",
			style.TitleStyle.Render(message.Timestamp.Format(time.TimeOnly)),
			style.TitleStyle.Render(message.Name),
			message.Message))
	}

	m.lastChatTimestamp = chatMessages[len(chatMessages)-1].Timestamp
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
	}

	task := resp.Result().(*api.TaskResponse)

	context.RunningTask = task
}
