package mailbox

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen"
	"farental/screen/generic/selectionlist"
	"farental/widget/mailboxlistitem"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
)

type Screen struct {
	selectionlist.Screen[api.MailBasicResponse]

	selectedMail *api.MailBasicResponse
}

func New() *Screen {
	s := new(Screen)

	s.Screen = selectionlist.New(lokyn.L("Mailbox"),
		mailboxlistitem.Constructor, s.loadData, s.submit)

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
	s.Screen.SetTitle(lokyn.L("Mailbox"))

	bubblehelp.SwitchContext(keybind.ContextFilterSelectionListWithNew)

	orvyn.SetPreviousScreen(screen.IDDashBoard)

	return nil
}

func (s *Screen) OnExit() any {
	if s.selectedMail != nil {
		return s.selectedMail
	}

	return nil
}

func (s *Screen) loadData() {
	var mails []api.MailBasicResponse

	res, err := helper.Fetch[[]api.MailBasicResponse](request.MailGetAll())

	if err != nil {
		s.SetStatusError(err)
		return
	}

	mails = *res

	s.SetItems(mails)
}

func (s *Screen) submit() bool {
	s.SetSubmitScreenID(screen.IDMailReader)
	s.selectedMail = s.getSelectedMail()

	return true
}

func (s *Screen) getSelectedMail() *api.MailBasicResponse {
	item := s.GetSelectedItem()

	return &item
}
