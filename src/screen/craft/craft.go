package craft

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen/generic/selectionlist"
	"farental/widget/craftlistitem"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"log"
)

type Screen struct {
	selectionlist.Screen[craftlistitem.Data]
}

func New() *Screen {
	s := new(Screen)

	s.Screen = selectionlist.New(lokyn.L("Crafts"), craftlistitem.Constructor,
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
	var data []craftlistitem.Data

	resp, err := helper.SendRequest(request.CraftGetAvailable())

	if err != nil {
		s.SetStatusError(err)
		return
	}

	crafts = *resp.Result().(*[]api.RecipeResponse)

	for _, c := range crafts {
		item := craftlistitem.Data{
			RecipeResponse: c,
			Amount:         0,
		}

		data = append(data, item)
	}

	s.SetItems(data)

	return
}

func (s *Screen) submit() bool {
	i := s.GetSelectedItem()

	req := request.CraftStart(i.ID, i.Amount)

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
