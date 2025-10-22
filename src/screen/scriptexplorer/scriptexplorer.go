package scriptexplorer

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen"
	"farental/screen/generic/selectionlist"
	"farental/widget/scriptexplorerlistitem"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-resty/resty/v2"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
)

type viewType uint8

const (
	own viewType = iota
	public
)

type Screen struct {
	selectionlist.Screen[api.ScriptBasicResponse]

	titleOwn    string
	titlePublic string

	newScript bool

	viewType viewType
}

func New() *Screen {
	s := new(Screen)

	s.titleOwn = lokyn.L("My scripts (%d/%d)")
	s.titlePublic = lokyn.L("Public scripts")

	s.viewType = own

	s.Screen = selectionlist.New(
		s.titleOwn,
		scriptexplorerlistitem.Constructor,
		s.loadScripts,
		s.submit,
	)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	s.Screen.OnEnter(i)

	s.viewType = own
	s.updateOwnTitle()

	bubblehelp.SwitchContext(keybind.ContextScriptExplorer)

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

		case key.Matches(m, keybind.EKey):
			s.newScript = false
			return orvyn.SwitchScreen(screen.IDScriptEditor)

		case key.Matches(m, keybind.Tab):
			s.switchViewType()
			s.loadScripts()
			s.FocusFirst()

			return nil
		}
	}

	return cmd
}

func (s *Screen) loadScripts() {
	var scripts []api.ScriptBasicResponse
	var req *resty.Request

	switch s.viewType {
	case own:
		req = request.ScriptGetOwn()
	case public:
		req = request.ScriptGetPublic()
	}

	resp, err := helper.SendRequest(req)

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

func (s *Screen) updateOwnTitle() {
	var count api.ScriptCountResponse

	resp, err := helper.SendRequest(request.ScriptGetCount())

	if err != nil {
		s.SetStatusError(err)
		count = api.ScriptCountResponse{
			Current: 0,
			Max:     0,
		}
	} else {
		count = *resp.Result().(*api.ScriptCountResponse)
	}

	title := fmt.Sprintf(s.titleOwn, count.Current, count.Max)

	s.SetTitle(title)
}

func (s *Screen) switchViewType() {
	switch s.viewType {
	case own:
		s.viewType = public
		s.SetTitle(s.titlePublic)
	case public:
		s.viewType = own
		s.updateOwnTitle()
	}
}
