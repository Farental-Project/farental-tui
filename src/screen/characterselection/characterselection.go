package characterselection

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/internal/orvyn/layout"
	"farental/screen"
	"farental/style"
	"farental/widget/help"
	"farental/widget/list"
	"farental/widget/statusmessage"
	"github.com/charmbracelet/bubbles/key"
	tealist "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/spf13/viper"
	"net/http"
)

type Screen struct {
	orvyn.BaseScreen

	title *orvyn.SimpleRenderable

	characters []tealist.Item
	list       *list.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	s.title = orvyn.NewSimpleRenderable(
		style.TitleStyle.Render(lang.L("Character selection")),
	)

	s.list = list.New(
		CharacterItemDelegate{},
		[]tealist.Item{},
	)

	s.list.PreferredSize.Width = style.LayoutWidth - 2 // items border
	s.list.MinSize.Height = 13

	s.statusMessage = statusmessage.New()
	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewVBoxFullLayout(orvyn.NewSize(10, 4), 2,
			[]orvyn.Renderable{
				s.title,
				orvyn.VGap,
				s.list,
				s.statusMessage,
				orvyn.VGap,
				s.help,
			},
		),
	)
	return s
}

func (s *Screen) OnEnter(_ interface{}) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextCharacterSel)

	s.loadCharacters()
	s.list.Select(0)

	return nil
}

func (s *Screen) OnExit() interface{} {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return tea.Quit

		case key.Matches(msg, keybind.Help):
			bubblehelp.ShowAll = !bubblehelp.ShowAll

			return nil

		case key.Matches(msg, keybind.Enter):
			ok := s.submit()

			if ok {
				return orvyn.SwitchScreen(screen.IDDashBoard)
			}

			return nil

		case key.Matches(msg, keybind.NKey):
			orvyn.SwitchScreen(screen.IDCharacterCreation)

		case key.Matches(msg, keybind.Esc):
			// Logout - Clear the logintoken in config and clear the client cookie.
			viper.Set("logintoken", "")

			viper.WriteConfig()

			context.Client.Cookies = make([]*http.Cookie, 0)

			return orvyn.SwitchScreen(screen.IDLogin)
		}
	}

	s.list.Update(msg)

	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) submit() bool {
	selectedItem, ok := s.list.SelectedItem().(CharacterItem)

	if !ok {
		return false
	}

	if selectedItem.ID == 0 {
		return false
	}

	req := request.CharacterSetActive(selectedItem.ID)

	resp, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return false
	}

	if resp.StatusCode() != http.StatusOK {
		return false
	}

	context.CharacterID = selectedItem.ID

	return true
}

// loadCharacters loads all the characters available for the current user.
func (s *Screen) loadCharacters() {
	var characters *[]api.CharacterBasicResponse
	var ok bool

	resp, err := helper.SendRequest(request.CharacterGetAll())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	characters, ok = resp.Result().(*[]api.CharacterBasicResponse)

	if !ok {
		return
	}

	s.characters = make([]tealist.Item, 0)

	for _, character := range *characters {
		s.characters = append(s.characters, NewCharacterItem(&character))
	}

	s.list.SetItems(s.characters)
}
