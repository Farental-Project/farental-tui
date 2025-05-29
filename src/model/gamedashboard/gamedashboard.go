package gamedashboard

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/model"
	"farental/model/widget/charactervitalinfo"
	"farental/model/widget/locationinfo"
	"farental/model/widget/simplelogviewer"
	"farental/style"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-resty/resty/v2"
	"log"
	"time"
)

var (
	styleDashboard = lipgloss.NewStyle().Width(100).AlignHorizontal(lipgloss.Center)
)

type Model struct {
	CharacterVitalInfo  charactervitalinfo.Model
	LocationInfo        locationinfo.Model
	EventLogViewer      simplelogviewer.Model
	ChatViewer          simplelogviewer.Model
	CharactersConnected simplelogviewer.Model

	lastEventLogTimestamp time.Time
	lastChatTimestamp     time.Time
}

func New() Model {
	return Model{
		CharacterVitalInfo:  charactervitalinfo.New(),
		LocationInfo:        locationinfo.New(),
		EventLogViewer:      simplelogviewer.New(),
		ChatViewer:          simplelogviewer.New(),
		CharactersConnected: simplelogviewer.New(),
	}
}

type tickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(30, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(model.InitCmd, doTick())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case tickMsg:
		m.UpdateData()

		return m, doTick()
	case model.InitMsg:

	}

	context.ContentManager.Update(msg)

	return m, nil
}

func (m Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		m.CharacterVitalInfo.View(),
		m.LocationInfo.View(),
		m.EventLogViewer.View(),
		m.ChatViewer.View(),
		m.CharactersConnected.View())
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
			style.TitleStyle.Render(message.Name),
			style.TitleStyle.Render(message.Timestamp.Format(time.TimeOnly)),
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

	if len(characters) == 0 {
		return
	}

	str = make([]string, len(characters))

	for _, character := range characters {
		str = append(str,
			style.TitleStyle.Render(fmt.Sprintf("%s %s - %s\n  %s",
				character.FirstName, character.LastName,
				character.RaceName, character.CurrentActivityTitle)))
	}

	m.CharactersConnected.SetContent(str)
}
