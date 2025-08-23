package mailbox

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen"
	"farental/screen/generic/selectionlist"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
)

type Screen struct {
	selectionlist.Screen

	selectedMail *api.MailBasicResponse
}

func New() *Screen {
	s := new(Screen)

	s.Screen = selectionlist.New(lokyn.L("Mailbox"),
		ListItemDelegate{}, s.loadData, s.submit)

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
	var items []list.Item

	items = make([]list.Item, 0)

	resp, err := helper.SendRequest(request.MailGetAll())

	if err != nil {
		s.SetStatusError(err)
		return
	}

	mails = *resp.Result().(*[]api.MailBasicResponse)

	for _, a := range mails {
		item := NewListItem(a)

		items = append(items, item)
	}

	s.SetItems(items)
}

func (s *Screen) submit() bool {
	s.SetSubmitScreenID(screen.IDMailReader)
	s.selectedMail = s.getSelectedMail()

	return true
}

func (s *Screen) getSelectedMail() *api.MailBasicResponse {
	item, ok := s.GetSelectedItem().(ListItem)

	if !ok {
		return nil
	}

	return &item.Mail
}
