package activity

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen/generic/selectionlist"
	"farental/widget/activitylistitem"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
)

type Screen struct {
	selectionlist.Screen[activitylistitem.Data]
}

func New() *Screen {
	s := new(Screen)

	s.Screen = selectionlist.New(lokyn.L("Activities"), activitylistitem.Constructor,
		s.loadActivities, s.submit)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	s.Screen.OnEnter(i)

	bubblehelp.SwitchContext(keybind.ContextFilterSelectionListIncDec)

	return nil
}

func (s *Screen) loadActivities() {
	var activities []api.ActivityResponse

	resp, err := helper.SendRequest(request.ActivityGetAvailable())

	if err != nil {
		s.SetStatusError(err)
		return
	}

	activities = *resp.Result().(*[]api.ActivityResponse)

	data := make([]activitylistitem.Data, 0)

	for _, a := range activities {
		data = append(data, activitylistitem.Data{
			ActivityResponse: a,
			DurationIndex:    0,
		})
	}

	s.SetItems(data)
}

func (s *Screen) submit() bool {
	var durationID uint

	i := s.GetSelectedItem()

	durationID = 0

	if len(i.Duration.Durations) > 1 {
		durationID = i.Duration.Durations[i.DurationIndex].ID
	} else {
		durationID = i.Duration.Durations[0].ID
	}

	req := request.ActivityStart(i.ID, durationID)

	_, err := helper.SendRequest(req)

	if err != nil {
		s.SetStatusError(err)
		return false
	}

	return true
}
