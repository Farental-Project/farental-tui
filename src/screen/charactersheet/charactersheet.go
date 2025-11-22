package charactersheet

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/widget/characteractivescript"
	"farental/widget/characterinfo"
	"farental/widget/equipmentsummary"
	"farental/widget/help"
	"farental/widget/skillssummary"
	"farental/widget/statssummary"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	characterInfo *characterinfo.Widget

	characterActiveScript *characteractivescript.Widget

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
	s.title.Style = orvyn.GetTheme().Style(theme.TitleStyleID)

	s.characterInfo = characterinfo.New()
	s.characterActiveScript = characteractivescript.New()
	s.equipmentSummary = equipmentsummary.New()
	s.statsSummary = statssummary.New()
	s.skillsSummary = skillssummary.New()
	s.statusMessage = statusmessage.New()
	s.help = help.New()

	elements := make([]layout.FixedRatioRenderable, 2)
	elements[0] = layout.NewFixedRatioRenderable(0.30, s.statsSummary)
	elements[1] = layout.NewFixedRatioRenderable(0.70, s.skillsSummary)

	s.statsSkillLayout = layout.NewHBoxFixedRatioLayout(0, 1,
		0,
		elements...,
	)

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(35,
			orvyn.GetTheme().Size(ftheme.LayoutWidthSizeID),
			10,
			s.title,
			orvyn.VGap,
			s.characterInfo,
			s.characterActiveScript,
			s.equipmentSummary,
			s.statsSkillLayout,
			s.statusMessage,
			s.help,
		),
	)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextCharacterSheet)

	s.updateData()

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

	s.characterActiveScript.UpdateData()

	s.equipmentSummary.UpdateData()

	s.statsSummary.UpdateData(characterInfo)

	s.skillsSummary.UpdateData(characterInfo)
}
