package scriptexplorer

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen"
	"farental/screen/dashboard"
	"farental/screen/dialog/popup"
	"farental/screen/generic/selectionlist"
	"farental/widget/scriptexplorerlistitem"
	"fmt"
	"net/http"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-resty/resty/v2"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
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

	newScript       bool
	duplicateScript bool
	selectScript    bool
	selectWarning   string

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

	s.newScript = false
	s.duplicateScript = false

	s.viewType = own
	s.updateOwnTitle()

	bubblehelp.SwitchContext(keybind.ContextScriptExplorer)

	orvyn.SetPreviousScreen(screen.IDDashBoard)

	s.loadScripts()

	return nil
}

func (s *Screen) OnExit() any {
	switch {
	case s.newScript:
		return nil
	case s.selectScript:
		if s.selectWarning != "" {
			statusMessage := dashboard.StatusMessageParam{
				Content: s.selectWarning,
				Type:    statusmessage.WarningMessage,
			}

			return statusMessage
		}
	}

	script := s.GetSelectedItem()

	if s.duplicateScript {
		script.IsDuplicated = true
	}

	return &script
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	cmd := s.Screen.Update(msg)

	if m, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(m, keybind.NKey):
			if s.GetFilteringState() != widgetlist.Filtering {
				s.newScript = true
				return orvyn.SwitchScreen(screen.IDScriptEditor)
			}

		case key.Matches(m, keybind.EKey):
			if s.GetFilteringState() != widgetlist.Filtering {
				s.newScript = false
				return orvyn.SwitchScreen(screen.IDScriptEditor)
			}

		case key.Matches(m, keybind.DKey):
			if s.GetFilteringState() != widgetlist.Filtering &&
				bubblehelp.IsKeybindVisible(keybind.DKey) {
				orvyn.OpenDialog("deleteConfirm", popup.NewYesNo(
					lokyn.L("Are you sure you want to delete the script ?"),
				), nil)
			}

		case key.Matches(m, keybind.CKey):
			if s.GetFilteringState() != widgetlist.Filtering {
				s.duplicateScript = true
				return orvyn.SwitchScreen(screen.IDScriptEditor)
			}

		case key.Matches(m, keybind.Tab):
			if s.GetFilteringState() != widgetlist.Filtering {
				s.switchViewType()
				s.loadScripts()
				s.FocusFirst()

				return nil
			}
		}
	}

	switch msg := msg.(type) {
	case orvyn.DialogExitMsg:
		switch msg.DialogID {
		case "deleteConfirm":
			val := msg.Param.(uint)

			switch val {
			case 1:
				s.deleteScript()
			default:
				return nil
			}
		}
	}

	return cmd
}

func (s *Screen) deleteScript() {
	script := s.GetSelectedItem()

	resp, err := helper.SendRequest(request.ScriptDelete(script.ID))

	if err != nil {
		return
	}

	if resp.StatusCode() == http.StatusOK {
		s.SetStatusSuccess(lokyn.L("Script successfully deleted !"))
		s.loadScripts()
	}
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
	selectedScript := s.GetSelectedItem()

	if len(selectedScript.ID) == 0 {
		return false
	}

	resp, err := helper.SendRequest(request.ScriptSetActive(selectedScript.ID))

	if err != nil {
		s.SetStatusError(err)
		return false
	}

	if resp.StatusCode() == http.StatusOK {
		response := *resp.Result().(*api.ScriptSetActiveResponse)

		s.selectScript = true
		s.selectWarning = response.WarningMessage

		return true
	}

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

		bubblehelp.SetKeybindVisible(keybind.DKey, false)
	case public:
		s.viewType = own
		s.updateOwnTitle()

		bubblehelp.SetKeybindVisible(keybind.DKey, true)
	}
}
