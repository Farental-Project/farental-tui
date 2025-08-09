package mailbox

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/screen"
	"farental/screen/generic/selectionlist"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
)

type Screen struct {
	selectionlist.Screen

	selectedMail *api.MailBasicResponse
}

func New() *Screen {
	s := new(Screen)

	s.Screen = selectionlist.New(lang.L("Mailbox"),
		ListItemDelegate{}, s.loadData, s.submit)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
	s.Screen.OnEnter(i)

	bubblehelp.SwitchContext(keybind.ContextFilterSelectionListWithNew)

	orvyn.SetPreviousScreen(screen.IDDashBoard)

	return nil
}

func (s *Screen) OnExit() interface{} {
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
