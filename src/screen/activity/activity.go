package activity

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen/generic/selectionlist"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
)

type Screen struct {
	selectionlist.Screen
}

func New() *Screen {
	s := new(Screen)

	s.Screen = selectionlist.New(lokyn.L("Activities"), ListItemDelegate{},
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
	var items []list.Item

	items = make([]list.Item, 0)

	resp, err := helper.SendRequest(request.ActivityGetAvailable())

	if err != nil {
		s.SetStatusError(err)
		return
	}

	activities = *resp.Result().(*[]api.ActivityResponse)

	for _, a := range activities {
		item := ListItem{
			Activity:      a,
			DurationIndex: 0,
		}

		items = append(items, item)
	}

	s.SetItems(items)
}

func (s *Screen) submit() bool {
	var durationID uint

	i, ok := s.GetSelectedItem().(ListItem)

	if !ok {
		return false
	}

	durationID = 0

	if len(i.Activity.Duration.Durations) > 0 {
		durationID = i.Activity.Duration.Durations[i.DurationIndex].ID
	} else {
		durationID = i.Activity.Duration.Durations[0].ID
	}

	req := request.ActivityStart(i.Activity.ID, durationID)

	_, err := helper.SendRequest(req)

	if err != nil {
		s.SetStatusError(err)
		return false
	}

	return true
}
