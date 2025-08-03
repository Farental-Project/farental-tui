package travel

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/lang"
	"farental/model"
	"farental/screen/generic/selectionlist"
	tealist "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
)

type Screen struct {
	selectionlist.Screen
}

func New() *Screen {
	s := new(Screen)

	s.Screen = selectionlist.New(lang.L("Travels"), ItemDelegate{},
		s.loadTravels, s.submit)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
	s.Screen.OnEnter(i)

	bubblehelp.SwitchContext(model.ContextFilterSelectionListBasic)

	return nil
}

func (s *Screen) submit() bool {
	i, ok := s.GetSelectedItem().(Item)

	if !ok {
		return false
	}

	req := request.TravelStart(i.ID)

	resp, err := helper.SendRequest(req)

	if err != nil {
		s.SetStatusError(err)
		return false
	}

	if resp.StatusCode() != 200 {
		return false
	}

	return true
}

func (s *Screen) loadTravels() {
	var travels *[]api.TravelResponse
	var ok bool

	resp, err := helper.SendRequest(request.TravelGetAvailable())

	if err != nil {
		s.SetStatusError(err)
		return
	}

	travels, ok = resp.Result().(*[]api.TravelResponse)

	if !ok {
		return
	}

	items := make([]tealist.Item, 0)

	for _, t := range *travels {
		items = append(items, NewItem(&t))
	}

	s.SetItems(items)
}
