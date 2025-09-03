package fight

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen/generic/selectionlist"
	"farental/widget/fightlistitem"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"log"
)

type Screen struct {
	selectionlist.Screen[fightlistitem.Data]
}

func New() *Screen {
	s := new(Screen)

	s.Screen = selectionlist.New(lokyn.L("Fights"), fightlistitem.Constructor,
		s.loadFights, s.submit)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	s.Screen.OnEnter(i)

	bubblehelp.SwitchContext(keybind.ContextFilterSelectionListPage)

	return nil
}

func (s *Screen) loadFights() {
	var fights []api.FightCompositionResponse

	data := make([]fightlistitem.Data, 0)

	resp, err := helper.SendRequest(request.FightGetAvailable())

	if err != nil {
		s.SetStatusError(err)
		return
	}

	fights = *resp.Result().(*[]api.FightCompositionResponse)

	for _, f := range fights {
		item := fightlistitem.Data{
			FightCompositionResponse: f,
			TotalPower:               0,
		}

		data = append(data, item)
	}

	s.SetItems(data)
}

func (s *Screen) submit() bool {
	i := s.GetSelectedItem()

	req := request.FightStart(i.ID)

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
