package characterselection

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/screen"
	"farental/widget/characterbasiclistitem"
	"farental/widget/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/spf13/viper"
	"net/http"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	list *list.Widget[api.CharacterBasicResponse]

	statusMessage *statusmessage.Widget

	help *help.Widget

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable(
		t.Style(theme.TitleStyleID).Render(lokyn.L("Character selection")),
	)

	s.list = list.New(characterbasiclistitem.Constructor)

	s.list.PreferredSize.Width = t.Size(ftheme.LayoutWidthSizeID)
	s.list.MinSize.Height = 13

	s.statusMessage = statusmessage.New()
	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(10, 4), 2,
			[]orvyn.Renderable{
				s.title,
				orvyn.VGap,
				s.list,
				s.statusMessage,
				s.help,
			},
		),
	)
	return s
}

func (s *Screen) OnEnter(_ any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextCharacterSel)

	s.loadCharacters()
	s.list.FocusFirst()

	return nil
}

func (s *Screen) OnExit() any {
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
	selectedItem := s.list.GetSelectedItem()

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
	var characters []api.CharacterBasicResponse

	resp, err := helper.SendRequest(request.CharacterGetAll())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	characters = *resp.Result().(*[]api.CharacterBasicResponse)

	if characters == nil {
		return
	}

	s.list.SetItems(characters)
}
