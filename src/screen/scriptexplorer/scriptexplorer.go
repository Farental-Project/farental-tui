package scriptexplorer

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen"
	"farental/screen/generic/selectionlist"
	"farental/widget/scriptexplorerlistitem"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
)

type Screen struct {
	selectionlist.Screen[api.ScriptBasicResponse]
}

func New() *Screen {
	s := new(Screen)

	s.Screen = selectionlist.New(
		lokyn.L("Scripts"),
		scriptexplorerlistitem.Constructor,
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

	resp, err := helper.SendRequest(request.ScriptGetPrivate())

	if err != nil {
		s.SetStatusError(err)
		return
	}

	scripts = *resp.Result().(*[]api.ScriptBasicResponse)

	s.SetItems(scripts)
}

func (s *Screen) submit() bool {
	return false
}
