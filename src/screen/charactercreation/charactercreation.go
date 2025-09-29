package charactercreation

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/widget/help"
	"farental/widget/multivalueselector"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/textinput"
)

type RaceData struct {
	api.RaceResponse
}

func (r RaceData) RenderValue() string {
	return r.Name
}

type Screen struct {
	title *orvyn.SimpleRenderable

	tiFirstname *textinput.Widget
	tiLastname  *textinput.Widget
	mvsRace     *multivalueselector.Widget[RaceData]

	raceDescription *orvyn.SimpleRenderable

	statusMessage *statusmessage.Widget

	help *help.Widget

	layout *layout.CenterLayout

	focusManager *orvyn.FocusManager
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable(
		t.Style(theme.TitleStyleID).Render(lokyn.L("New character")),
	)

	s.tiFirstname = textinput.New()
	s.tiFirstname.Placeholder = lokyn.L("First name")

	s.tiLastname = textinput.New()
	s.tiLastname.Placeholder = lokyn.L("Last name")

	s.mvsRace = multivalueselector.New[RaceData]()
	s.mvsRace.OnBlur()

	s.raceDescription = orvyn.NewSimpleRenderable("")
	s.raceDescription.Style = t.Style(theme.DimSecondaryTextStyleID).
		AlignHorizontal(lipgloss.Center)
	s.raceDescription.SizeConstraint = true

	s.statusMessage = statusmessage.New()

	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxLayout(20,
			[]orvyn.Renderable{
				s.title,
				orvyn.VGap,
				s.tiFirstname,
				s.tiLastname,
				s.mvsRace,
				s.raceDescription,
				orvyn.VGap,
				s.statusMessage,
				s.help,
			},
		),
	)

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.tiFirstname)
	s.focusManager.Add(s.tiLastname)
	s.focusManager.Add(s.mvsRace)
	s.focusManager.Focus(0)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextCharacterCreation)

	s.tiFirstname.SetValue("")
	s.tiLastname.SetValue("")
	s.mvsRace.SetSelected(0)

	s.focusManager.Focus(0)

	s.loadRaces()

	s.raceDescription.SetValue(
		s.mvsRace.GetSelectedValue().Description)

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

		case key.Matches(msg, keybind.Esc):
			return orvyn.SwitchToPreviousScreen()

		case key.Matches(msg, keybind.Enter):
			ok := s.submit()

			if ok {
				return orvyn.SwitchToPreviousScreen()
			}

			return nil
		}
	}

	s.focusManager.Update(msg)

	s.raceDescription.SetValue(
		s.mvsRace.GetSelectedValue().Description)

	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) submit() bool {
	req := request.CharacterCreate()

	req.SetBody(
		api.CharacterCreateBody{
			FirstName: s.tiFirstname.Value(),
			LastName:  s.tiLastname.Value(),
			RaceID:    s.mvsRace.GetSelectedValue().ID,
		})

	_, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return false
	}

	return true

}

func (s *Screen) loadRaces() {
	resp, err := helper.SendRequest(request.DataGetAllRace())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	races := *resp.Result().(*[]api.RaceResponse)

	raceValues := make(map[string]RaceData, len(races))
	keys := make([]string, len(races))

	for k, v := range races {
		keys[k] = v.Name
		raceValues[keys[k]] = RaceData{
			v,
		}
	}

	s.mvsRace.SetValues(keys, raceValues)
	s.mvsRace.SetSelected(0)
}
