package craft

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
	"log"
)

type Screen struct {
	selectionlist.Screen
}

func New() *Screen {
	s := new(Screen)

	s.Screen = selectionlist.New(lokyn.L("Crafts"), ListItemDelegate{},
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
	var items []list.Item

	items = make([]list.Item, 0)

	resp, err := helper.SendRequest(request.CraftGetAvailable())

	if err != nil {
		s.SetStatusError(err)
		return
	}

	crafts = *resp.Result().(*[]api.RecipeResponse)

	for _, c := range crafts {
		item := NewListItem(&c)

		items = append(items, item)
	}

	s.SetItems(items)

	return
}

func (s *Screen) submit() bool {
	i, ok := s.GetSelectedItem().(ListItem)

	if !ok {
		return false
	}

	req := request.CraftStart(i.CraftRecipe.ID, i.Amount)

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
