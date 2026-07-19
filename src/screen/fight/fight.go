package fight

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen"
	"farental/screen/dialog/fightinspector"
	"farental/screen/generic/selectionlist"
	"farental/widget/characteractivescript"
	"farental/widget/characterinfo"
	"farental/widget/fightlistitem"
	"log"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
)

type Screen struct {
	selectionlist.Screen[fightlistitem.Data]

	characterInfo         *characterinfo.Widget
	characterActiveScript *characteractivescript.Widget
}

func New() *Screen {
	s := new(Screen)

	s.characterInfo = characterinfo.New()
	s.characterActiveScript = characteractivescript.New()

	headerLayout := layout.NewMaxWidthVBoxLayout(0,
		s.characterInfo,
		s.characterActiveScript)

	s.Screen = selectionlist.NewWithHeader(lokyn.L("Fights"), fightlistitem.Constructor,
		s.loadFights, s.submit, headerLayout)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	s.Screen.OnEnter(i)

	s.Screen.SetTitle(lokyn.L("Fights"))

	orvyn.SetPreviousScreen(screen.IDDashBoard)

	bubblehelp.SwitchContext(keybind.ContextFightList)

	s.updateData()

	return nil
}

func (s *Screen) updateData() {
	characterInfo, currency, err := context.RefreshCharacterInfo(false)

	if err != nil {
		s.Screen.SetStatusError(err)
		return
	}

	data := characterinfo.ConvertCharacterInfoResponseToData(characterInfo, currency)
	s.characterInfo.UpdateData(data)

	s.characterActiveScript.UpdateData()
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	cmd := s.Screen.Update(msg)

	if m, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(m, keybind.IKey):
			orvyn.OpenDialog("fightInspector",
				fightinspector.New(s.Screen.GetSelectedItem().Actors), nil)

			return nil
		case key.Matches(m, keybind.HKey):
			return orvyn.SwitchScreen(screen.IDFightHistory)
		}
	}

	switch msg := msg.(type) {
	case orvyn.DialogExitMsg:
		if msg.DialogID == "fightInspector" {
			bubblehelp.SwitchContext(keybind.ContextFightList)
		}
	}

	return cmd
}

func (s *Screen) loadFights() {
	var fights []api.FightCompositionResponse

	data := make([]fightlistitem.Data, 0)

	res, err := helper.Fetch[[]api.FightCompositionResponse](request.FightGetAvailable())

	if err != nil {
		s.SetStatusError(err)
		return
	}

	fights = *res

	for _, f := range fights {
		item := fightlistitem.Data{
			FightCompositionResponse: f,
			TotalPower:               0,
			Amount:                   1,
		}

		data = append(data, item)
	}

	s.SetItems(data)
}

func (s *Screen) submit() bool {
	i := s.GetSelectedItem()

	req := request.FightStart(i.ID, i.Amount)

	resp, err := helper.SendRequest(req)

	if err != nil {
		s.SetStatusError(err)
		return false
	}

	if resp.StatusCode() != 200 {
		log.Println(resp.Error())
		return false
	}

	return true
}
