package fight

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/lang"
	"farental/model"
	"farental/screen/generic/selectionlist"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"log"
)

type Screen struct {
	selectionlist.Screen
}

func New() *Screen {
	s := new(Screen)

	s.Screen = selectionlist.New(lang.L("Fights"), ListItemDelegate{},
		s.loadFights, s.submit)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
	s.Screen.OnEnter(i)

	bubblehelp.SwitchContext(model.ContextFilterSelectionListPage)

	return nil
}

func (s *Screen) loadFights() {
	var fights []api.FightCompositionResponse
	var items []list.Item

	items = make([]list.Item, 0)

	resp, err := helper.SendRequest(request.FightGetAvailable())

	if err != nil {
		s.SetStatusError(err)
		return
	}

	fights = *resp.Result().(*[]api.FightCompositionResponse)

	for _, f := range fights {
		item := NewListItem(f)

		items = append(items, item)
	}

	s.SetItems(items)
}

func (s *Screen) submit() bool {
	i, ok := s.GetSelectedItem().(ListItem)

	if !ok {
		return false
	}

	req := request.FightStart(i.FightCompo.ID)

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
