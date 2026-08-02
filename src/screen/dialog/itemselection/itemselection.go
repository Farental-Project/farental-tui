package itemselection

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen/generic/selectionlist"
	"farental/widget/inventorylistitem"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Screen struct {
	selectionlist.Screen[api.StackResponse]

	submitted bool
}

var _ orvyn.Screen = (*Screen)(nil)

func New() *Screen {
	s := new(Screen)

	s.Screen = selectionlist.New(lokyn.L("Items"),
		inventorylistitem.Constructor, s.loadData, s.submit)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	cmd := s.Screen.OnEnter(i)

	s.SetTitle(lokyn.L("Items"))

	bubblehelp.SwitchContext(keybind.ContextFilterSelectionListBasic)

	return cmd
}

func (s *Screen) OnExit() any {
	if s.submitted {
		return s.GetSelectedItem()
	}

	return nil
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

func (s *Screen) loadData() {
	inventory, err := helper.Fetch[api.InventoryResponse](request.InventoryGetShareable())

	if err != nil {
		return
	}

	s.SetItems(inventory.CreateGroupedStackResponseList())
}

func (s *Screen) submit() bool {
	return true
}
