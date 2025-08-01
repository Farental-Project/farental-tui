package dashboard

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
func (s *Screen) updateErr(err *error) {
	if *err != nil {
		s.statusMessage.SetError(*err)
	}
}

func (s *Screen) updateData() {
	var err error

	defer s.updateErr(&err)

	resp, err := helper.SendRequest(request.CharacterGetInfo())

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

	resp, err = helper.SendRequest(
		request.CharacterGetCurrencyAmount(api.Grynars))

	if err != nil {
		return
	}

	currencyResp := resp.Result().(*api.CurrencyResponse)

	s.characterInfo.UpdateData(characterInfo, currencyResp.Amount)
	s.locationInfo.UpdateData(&characterInfo.Location)
	s.updateRunningTask()
	s.updateEventLog()
	s.updateChat()
	s.updateVisibleCharacters()
}

func (s *Screen) updateEventLog() {
	var req *resty.Request
	var queryParam string

	req = request.CharacterGetEventLog()

	queryParam = ""
	length := len(s.logEvent.GetContent())

	if length > 0 {
		queryParam = s.lastEventLogTimestamp.Format(time.DateTime)
	}

	req.SetQueryParam("lastTimestamp", queryParam)

	resp, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	eventLog := resp.Result().(*api.EventLogResponse)

	if len(eventLog.Entries) == 0 {
		return
	}

	for _, entry := range eventLog.Entries {
		s.logEvent.AppendContent(fmt.Sprintf("%s - %s",
			style.TitleStyle.Render(
				entry.Timestamp.Format(viper.GetString("datetimeformat"))),
			entry.Value,
		))
	}

	s.lastEventLogTimestamp = eventLog.Entries[len(eventLog.Entries)-1].Timestamp
}

func (s *Screen) updateChat() {
	context.UpdateChat()

	if len(context.ChatContent) > len(s.logChat.GetContent()) {
		s.logChat.SetContent(context.ChatContent)
	}
}

func (s *Screen) updateVisibleCharacters() {
	var str []string

	resp, err := helper.SendRequest(request.LocationGetCharacters())

	if err != nil {
		s.statusMessage.SetError(err)
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

	s.logCharacters.SetContent(str)
}

func (s *Screen) updateRunningTask() {
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
