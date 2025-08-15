package charactersheet

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/orvyn"
	layout "farental/layout"
	"farental/style"
	"farental/widget/characterinfo"
	"farental/widget/equipmentsummary"
	"farental/widget/help"
	"farental/widget/skillssummary"
	"farental/widget/statssummary"
	"farental/widget/statusmessage"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	characterInfo *characterinfo.Widget

	equipmentSummary *equipmentsummary.Widget

	statsSummary *statssummary.Widget

	skillsSummary *skillssummary.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	statsSkillLayout *layout.HBoxFixedRatio

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	s.title = orvyn.NewSimpleRenderable(lokyn.L("Character"))
	s.title.Style = style.TitleStyle

	s.characterInfo = characterinfo.New()
	s.equipmentSummary = equipmentsummary.New()
	s.statsSummary = statssummary.New()
	s.skillsSummary = skillssummary.New()
	s.statusMessage = statusmessage.New()
	s.help = help.New()

	s.statsSkillLayout = layout.NewHBoxFixedRatioLayout(0, 1,
		0,
		[]layout.FixedRatioRenderable{
			layout.NewFixedRatioRenderable(0.30, s.statsSummary),
			layout.NewFixedRatioRenderable(0.70, s.skillsSummary),
		},
	)

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(35,
			style.LayoutWidth,
			10,
			[]orvyn.Renderable{
				s.title,
				orvyn.VGap,
				s.characterInfo,
				s.equipmentSummary,
				s.statsSkillLayout,
				s.statusMessage,
				s.help,
			},
		),
	)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextCharacterSheet)

	s.updateData()

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

		case key.Matches(msg, keybind.Esc):
			return orvyn.SwitchToPreviousScreen()

		}
	}

	s.skillsSummary.Update(msg)

	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) updateData() {
	characterInfo := context.CharacterInfo

	req := request.CharacterGetCurrencyAmount(api.Grynars)

	resp, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	currencyResp := resp.Result().(*api.CurrencyResponse)

	s.characterInfo.UpdateData(characterInfo, currencyResp.Amount)

	s.equipmentSummary.UpdateData()

	s.statsSummary.UpdateData(characterInfo)

	s.skillsSummary.UpdateData(characterInfo)
}
