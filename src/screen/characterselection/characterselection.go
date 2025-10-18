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
	"net/http"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
)

type gotoDashboardMsg int

func gotoDashboardCmd() tea.Msg {
	return gotoDashboardMsg(1)
}

type gotoCharacterCreationMsg int

func gotoCharacterCreationCmd() tea.Msg {
	return gotoCharacterCreationMsg(1)
}

type Screen struct {
	title *orvyn.SimpleRenderable

	list *list.Widget[api.CharacterBasicResponse]

	statusMessage *statusmessage.Widget

	help *help.Widget

	layout *layout.CenterLayout

	noCharacters bool
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable(
		t.Style(theme.TitleStyleID).Render(lokyn.L("Character selection")),
	)

	s.list = list.New(characterbasiclistitem.Constructor)
	s.list.SetFilterable(false)

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
	var cmd tea.Cmd

	bubblehelp.SwitchContext(keybind.ContextCharacterSel)

	s.loadCharacters()

	if orvyn.GetPreviousScreen() != screen.IDDashBoard {
		if len(s.list.GetItems()) == 0 {
			s.noCharacters = true
			cmd = gotoCharacterCreationCmd
		} else {
			s.noCharacters = false
			resp, _ := helper.SendRequest(request.CharacterGetActive())

			if resp.StatusCode() == http.StatusOK {
				charInfo, ok := resp.Result().(*api.CharacterBasicResponse)

				if ok {
					context.CharacterID = charInfo.ID
					cmd = gotoDashboardCmd
				}
			}
		}
	}

	s.list.FocusFirst()

	return cmd
}

func (s *Screen) OnExit() any {
	context.Reset()
	return s.noCharacters // Only usefull in charactercreation screen. To set the esc to logout.
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
			context.Logout()

			return orvyn.SwitchScreen(screen.IDLogin)
		}
	case gotoDashboardMsg:
		return orvyn.SwitchScreen(screen.IDDashBoard)

	case gotoCharacterCreationMsg:
		return orvyn.SwitchScreen(screen.IDCharacterCreation)

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
