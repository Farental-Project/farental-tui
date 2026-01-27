package craft

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen/generic/skillgroupedselectionlist"
	"farental/widget/craftlistitem"
	"farental/widget/skillgrouplistitem"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
)

type Screen struct {
	skillgroupedselectionlist.Screen[craftlistitem.Data]
}

func New() *Screen {
	s := new(Screen)

	s.Screen = skillgroupedselectionlist.New(lokyn.L("Crafts"),
		craftlistitem.Constructor,
		s.loadCrafts, s.submit)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	s.Screen.OnEnter(i)

	bubblehelp.SwitchContext(keybind.ContextCraft)

	return nil
}

func (s *Screen) loadCrafts() {
	var crafts []api.RecipeResponse
	var groups []skillgrouplistitem.Data[craftlistitem.Data]

	resp, err := helper.SendRequest(request.CraftGetAvailable())

	if err != nil {
		s.SetStatusError(err)
		return
	}

	crafts = *resp.Result().(*[]api.RecipeResponse)
	currentSkillID := uint(0)
	currentGroupIndex := -1

	for _, c := range crafts {
		if currentSkillID != c.Skill.ID {
			currentSkillID = c.Skill.ID
			currentGroupIndex++

			group := skillgrouplistitem.Data[craftlistitem.Data]{
				Items:     make([]craftlistitem.Data, 0),
				SkillName: c.Skill.Name,
			}

			groups = append(groups, group)
		}

		item := craftlistitem.Data{
			RecipeResponse: c,
			CraftAmount:    0,
		}

		groups[currentGroupIndex].Items = append(groups[currentGroupIndex].Items, item)
	}

	s.SetItems(groups)
}

func (s *Screen) submit() bool {
	i := s.GetSelectedItem()

	if i.CraftAmount <= 0 {
		return false
	}

	req := request.CraftStart(i.ID, i.CraftAmount)

	resp, err := helper.SendRequest(req)

	if err != nil {
		s.SetStatusError(err)
		return false
	}

	if resp.StatusCode() != 200 {
		log.Println(resp.Error())
		return false
	}

	return true
}
