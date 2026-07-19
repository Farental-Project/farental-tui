package dashboard

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/style"
	"farental/widget/characterinfo"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/spf13/viper"
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

	characterInfo, currency, err := context.RefreshCharacterInfo(true)

	if err != nil {
		return
	}

	data := characterinfo.ConvertCharacterInfoResponseToData(characterInfo, currency)
	s.characterInfo.UpdateData(data)
	s.locationInfo.UpdateData(&characterInfo.Location)

	s.runningTask.RefreshCurrentCharacter()

	s.updateEventLog()
	s.updateChat()
	s.updateVisibleCharacters()
}

func (s *Screen) updateEventLog() {
	var req *resty.Request
	var queryParam string
	var firstInit bool

	req = request.CharacterGetEventLog()

	firstInit = true
	queryParam = ""
	length := len(s.logEvent.GetContent())

	if length > 0 {
		queryParam = s.lastEventLogTimestamp.UTC().Format(time.DateTime)
		firstInit = false
	}

	req.SetQueryParam("lastTimestamp", queryParam)

	eventLog, err := helper.Fetch[api.EventLogResponse](req)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	if len(eventLog.Entries) == 0 {
		return
	}

	ts := orvyn.GetTheme().Style(theme.TitleStyleID)

	format := viper.GetString("datetimeformat")

	if firstInit {
		var content []string

		for _, entry := range eventLog.Entries {
			log := fmt.Sprintf("%s - %s",
				ts.Render(
					entry.Timestamp.Local().Format(format)),
				entry.Value,
			)

			content = append(content, log)
		}

		s.logEvent.SetContent(content)
	} else {
		for _, entry := range eventLog.Entries {
			s.logEvent.AppendContent(fmt.Sprintf("%s - %s",
				ts.Render(
					entry.Timestamp.Local().Format(format)),
				entry.Value,
			))
		}
	}

	s.lastEventLogTimestamp = eventLog.Entries[len(eventLog.Entries)-1].Timestamp
}

func (s *Screen) updateChat() {
	context.UpdateChat()

	ctxContent := context.ChatContent
	content := s.logChat.GetContent()

	if len(ctxContent) == 0 {
		s.logChat.SetContent(ctxContent)
		return
	}

	if len(content) == 0 {
		s.logChat.SetContent(ctxContent)
		return
	}

	if ctxContent[len(ctxContent)-1] != content[len(content)-1] {
		s.logChat.SetContent(ctxContent)
	}
}

func (s *Screen) updateVisibleCharacters() {
	var str []string
	var b strings.Builder

	res, err := helper.Fetch[[]api.CharacterBasicWithActivityResponse](request.LocationGetCharacters())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	characters := *res
	t := orvyn.GetTheme()
	ds := t.Style(theme.DimTextStyleID)

	str = make([]string, 0)

	for _, character := range characters {
		b.Reset()

		raceStyle := style.RaceStyle(character.RaceName)

		b.WriteString(t.Style(theme.TitleStyleID).Render(
			fmt.Sprintf("%s %s", character.FirstName, character.LastName),
		))
		b.WriteString(ds.Render(" - "))
		b.WriteString(raceStyle.Render(character.RaceName))
		b.WriteString("\n  ")
		b.WriteString(ds.Render(character.CurrentActivityTitle))

		str = append(str, b.String())
	}

	s.logCharacters.SetContent(str)
}
