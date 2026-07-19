package characterinspector

import (
	"farental/core/data/api"
	"farental/internal/context"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/internal/ticker"
	"farental/widget/characterinfo"
	"farental/widget/equipmentsummary"
	"farental/widget/help"
	"farental/widget/runningtask"
	"farental/widget/skillssummary"
	"farental/widget/statssummary"
	"log"

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

	runningTask *runningtask.Widget

	ticker *ticker.Ticker

	characterInfo *characterinfo.Widget

	equipmentSummary *equipmentsummary.Widget

	statsSummary *statssummary.Widget

	skillsSummary *skillssummary.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	statsSkillLayout *layout.HBoxFixedRatio

	layout *layout.CenterLayout

	character *api.CharacterBasicResponse
}

func New() *Screen {
	s := new(Screen)

	s.title = orvyn.NewSimpleRenderable("Character inspector")
	s.title.Style = orvyn.GetTheme().Style(theme.TitleStyleID)

	s.characterInfo = characterinfo.New()
	s.runningTask = runningtask.New()
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
			s.runningTask,
			s.characterInfo,
			s.equipmentSummary,
			s.statsSkillLayout,
			s.statusMessage,
			s.help,
		),
	)

	s.ticker = ticker.New(60, func() {
		if err := context.RefreshRunningTask(); err != nil {
			log.Println(err)
		}
	})

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	// TODO : create basic context
	bubblehelp.SwitchContext(keybind.ContextCharacterSheet)

	bubblehelp.SetKeybindVisible(keybind.IKey, false)

	s.character = nil

	character, ok := i.(*api.CharacterBasicResponse)

	if !ok {
		return orvyn.SwitchToPreviousScreen()
	}

	s.character = character

	s.title.SetValue(lokyn.L("Character inspector"))

	s.statusMessage.Reset()

	s.updateData()

	if err := context.RefreshRunningTask(); err != nil {
		log.Println(err)
	}

	return tea.Batch(s.runningTask.Init(), s.ticker.Start())
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

	case orvyn.TickMsg:
		handled, cmd := s.ticker.Handle(msg)

		if !handled {
			return nil
		}

		return cmd
	}

	s.skillsSummary.Update(msg)

	return s.runningTask.Update(msg)
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) updateData() {
	data := characterinfo.ConvertCharacterBasicResponseToData(s.character)
	s.characterInfo.UpdateData(data)

	// TODO : Update call
	s.equipmentSummary.UpdateData()

	s.statsSummary.UpdateData(s.character.Stats)

	s.skillsSummary.UpdateData(s.character.Skills)
}
