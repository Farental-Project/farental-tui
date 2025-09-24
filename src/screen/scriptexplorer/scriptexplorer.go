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

	newScript bool
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

func (s *Screen) OnEnter(i any) tea.Cmd {
	s.Screen.OnEnter(i)

	bubblehelp.SwitchContext(keybind.ContextFilterSelectionListWithNew)

	orvyn.SetPreviousScreen(screen.IDDashBoard)

	return nil
}

func (s *Screen) OnExit() any {
	if s.newScript {
		return nil
	}

	script := s.GetSelectedItem()

	return &script
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	cmd := s.Screen.Update(msg)

	if m, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(m, keybind.NKey):
			s.newScript = true
			return orvyn.SwitchScreen(screen.IDScriptEditor)

		case key.Matches(m, keybind.Enter):
			s.newScript = false
			return orvyn.SwitchScreen(screen.IDScriptEditor)

		}
	}

	return cmd
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
