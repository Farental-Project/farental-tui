package abilityselection

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen/generic/selectionlist"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Screen struct {
	selectionlist.Screen[api.AbilityResponse]

	submitted          bool
	availableAbilities *[]api.AbilityResponse
}

func New() *Screen {
	s := new(Screen)

	s.Screen = selectionlist.New(lokyn.L("Abilities"),
		Constructor, s.loadData, s.submit)

	s.submitted = false

	return s
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if m, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(m, keybind.Enter):
			if s.GetFilteringState() != widgetlist.Filtering {
				if s.submit() {
					s.submitted = true
					return orvyn.CloseDialog()
				}
			}

		case key.Matches(m, keybind.Esc):
			if s.GetFilteringState() == widgetlist.Unfiltered {
				return orvyn.CloseDialog()
			}
		}
	}

	cmd := s.Screen.Update(msg)

	return cmd
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	availableAbilities, ok := i.(*[]api.AbilityResponse)

	if !ok {
		availableAbilities = nil
	}

	s.availableAbilities = availableAbilities

	cmd := s.Screen.OnEnter(i)

	s.Screen.SetTitle(lokyn.L("Abilities"))

	bubblehelp.SwitchContext(keybind.ContextScriptAbilitySelection)

	return cmd
}

func (s *Screen) OnExit() any {
	if s.submitted {
		return s.GetSelectedItem()
	}

	return nil
}

func (s *Screen) loadData() {
	var abilities []api.AbilityResponse

	if s.availableAbilities != nil {
		abilities = *s.availableAbilities
	} else {
		res, err := helper.Fetch[[]api.AbilityResponse](request.AbilityGetAll())

		if err != nil {
			s.SetStatusError(err)
			return
		}

		abilities = *res
	}

	s.SetItems(abilities)

}

func (s *Screen) submit() bool {
	return true
}
