package characterlocationlist

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen"
	"farental/screen/generic/selectionlist"
	"farental/widget/characterbasiclistitem"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
)

type Screen struct {
	selectionlist.Screen[characterbasiclistitem.Data]

	selectedCharacter *api.CharacterBasicResponse
}

func New() *Screen {
	s := new(Screen)

	s.Screen = selectionlist.New(lokyn.L("Characters around you"),
		characterbasiclistitem.Constructor, s.loadData, s.submit)

	return s
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	cmd := s.Screen.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.NKey):
			return orvyn.SwitchScreen(screen.IDMailEditor)
		}
	}

	return cmd
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	s.Screen.OnEnter(i)
	s.Screen.SetTitle(lokyn.L("Characters around you"))

	bubblehelp.SwitchContext(keybind.ContextFilterSelectionListBasic)

	orvyn.SetPreviousScreen(screen.IDDashBoard)

	return nil
}

func (s *Screen) OnExit() any {
	if s.selectedCharacter != nil {
		return s.selectedCharacter.ID
	}

	return nil
}

func (s *Screen) loadData() {
	var characters []characterbasiclistitem.Data

	res, err := helper.Fetch[[]api.CharacterBasicWithActivityResponse](request.LocationGetCharacters())

	if err != nil {
		s.SetStatusError(err)
		return
	}

	for _, c := range *res {
		characters = append(characters, characterbasiclistitem.Data{
			CharacterBasicResponse: c.CharacterBasicResponse,
			ShowLocation:           false,
		})
	}

	s.SetItems(characters)
}

func (s *Screen) submit() bool {
	s.SetSubmitScreenID(screen.IDCharacterInspector)
	s.selectedCharacter = s.getSelectedCharacter()

	return true
}

func (s *Screen) getSelectedCharacter() *api.CharacterBasicResponse {
	item := s.GetSelectedItem()

	return &item.CharacterBasicResponse
}
