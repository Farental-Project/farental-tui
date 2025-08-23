package scriptexplorer

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/orvyn"
	"farental/screen"
	"farental/screen/generic/selectionlist"
	"github.com/charmbracelet/bubbles/key"
	tealist "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
)

type Screen struct {
	selectionlist.Screen
}

func New() *Screen {
	s := new(Screen)

	s.Screen = selectionlist.New(
		lokyn.L("Scripts"),
		ListItemDelegate{},
		s.loadScripts,
		s.submit,
	)

	return s
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	cmd := s.Screen.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.NKey):
			return orvyn.SwitchScreen(screen.IDScriptEditor)
		}
	}

	return cmd
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	s.Screen.OnEnter(i)

	bubblehelp.SwitchContext(keybind.ContextFilterSelectionListWithNew)

	return nil
}

func (s *Screen) loadScripts() {
	var scripts []api.ScriptBasicResponse
	var items []tealist.Item

	items = make([]tealist.Item, 0)

	resp, err := helper.SendRequest(request.ScriptGetPrivate())

	if err != nil {
		s.SetStatusError(err)
		return
	}

	scripts = *resp.Result().(*[]api.ScriptBasicResponse)

	for _, s := range scripts {
		item := NewListItem(s)

		items = append(items, item)
	}

	s.SetItems(items)
}

func (s *Screen) submit() bool {
	return false
}
