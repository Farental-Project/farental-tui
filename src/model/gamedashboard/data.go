package gamedashboard

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/style"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
	"log"
	"time"
)

// updateErr used as defer to replace the error message only when necessary.
func (m *Model) updateErr(err *error) {
	if *err != nil {
		m.resetMsg()
		m.ErrMsg = *err
	}
}

func (m *Model) UpdateData() {
	var req *resty.Request
	var err error

	defer m.updateErr(&err)

	req = request.CharacterGetInfo()

	resp, err := helper.SendRequest(req)

	if err != nil {
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

	resp, err = helper.SendRequest(req)

	if err != nil {
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

	resp, err := helper.SendRequest(req)

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
				entry.Timestamp.Format(viper.GetString("datetimeformat"))),
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
	var str []string

	resp, err := helper.SendRequest(request.LocationGetCharacters())

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
	resp, err := helper.SendRequest(request.TaskGetRunning())

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
