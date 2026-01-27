package activity

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen/generic/skillgroupedselectionlist"
	"farental/widget/activitylistitem"
	"farental/widget/skillgrouplistitem"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
)

type Screen struct {
	skillgroupedselectionlist.Screen[activitylistitem.Data]
}

func New() *Screen {
	s := new(Screen)

	s.Screen = skillgroupedselectionlist.New(lokyn.L("Activities"), activitylistitem.Constructor,
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
	var groups []skillgrouplistitem.Data[activitylistitem.Data]

	resp, err := helper.SendRequest(request.ActivityGetAvailable())

	if err != nil {
		s.SetStatusError(err)
		return
	}

	activities = *resp.Result().(*[]api.ActivityResponse)
	currentSkillID := uint(0)
	currentGroupIndex := -1

	for _, a := range activities {
		if currentSkillID != a.Skill.ID {
			currentSkillID = a.Skill.ID
			currentGroupIndex++

			group := skillgrouplistitem.Data[activitylistitem.Data]{
				Items:     make([]activitylistitem.Data, 0),
				SkillName: a.Skill.Name,
			}

			groups = append(groups, group)
		}

		item := activitylistitem.Data{
			ActivityResponse: a,
			DurationIndex:    0,
		}

		groups[currentGroupIndex].Items = append(groups[currentGroupIndex].Items, item)
	}

	s.SetItems(groups)
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
