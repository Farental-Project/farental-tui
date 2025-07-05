package charactersheet

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/model"
	"farental/model/widget/charactervitalinfo"
	"farental/model/widget/equipmentsummary"
	"farental/model/widget/skillssummary"
	"farental/model/widget/statssummary"
	"farental/model/widget/widgetcontainer"
	"farental/style"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Data   api.CharacterInfoResponse
	ErrMsg error

	CharacterVitalInfo charactervitalinfo.Model

	EquipmentSummary          equipmentsummary.Model
	EquipmentSummaryContainer widgetcontainer.Model

	StatsSummary          statssummary.Model
	StatsSummaryContainer widgetcontainer.Model

	SkillSummary          skillssummary.Model
	SkillSummaryContainer widgetcontainer.Model
}

func New() Model {
	m := Model{
		CharacterVitalInfo: charactervitalinfo.New(style.LayoutWidth),
		EquipmentSummary:   equipmentsummary.New(style.LayoutWidth),
		StatsSummary:       statssummary.New(25),
		SkillSummary:       skillssummary.New(48),
	}

	m.EquipmentSummaryContainer = widgetcontainer.New(m.EquipmentSummary,
		lang.L("Equipment"), style.LayoutWidth, 6)

	m.StatsSummaryContainer = widgetcontainer.New(m.StatsSummary,
		lang.L("Stats"), 25, 11)

	m.SkillSummaryContainer = widgetcontainer.New(m.SkillSummary,
		lang.L("Skills"), 48, 25)

	return m
}

func (m Model) Init() tea.Cmd {
	return model.InitCmd
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	defer context.ContentManager.UpdateCurrentContent(m)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return m, tea.Quit
		case key.Matches(msg, keybind.Esc):
			return context.ContentManager.
				SwitchContent(m, model.ContentGameDashboard)
		}
	case model.InitMsg:
		m.UpdateData()

		context.KeymapManager.SwitchContext(model.ContextCharacterSheet)
	}

	context.ContentManager.Update(msg)

	return m, nil
}

func (m Model) View() string {
	skillStat := lipgloss.JoinHorizontal(lipgloss.Top,
		m.StatsSummaryContainer.View(),
		m.SkillSummaryContainer.View())

	tui := lipgloss.JoinVertical(lipgloss.Center,
		style.ContainerStyle.Render(m.CharacterVitalInfo.View()),
		m.EquipmentSummaryContainer.View(),
		skillStat,
		context.KeymapManager.View(style.LayoutWidth))

	return lipgloss.Place(
		context.ContentManager.ScreenWidth,
		context.ContentManager.ScreenHeight,
		lipgloss.Center,
		lipgloss.Center,
		tui)
}

func (m *Model) UpdateData() {
	resp, err := helper.SendRequest(request.CharacterGetInfo())

	if err != nil {
		m.ErrMsg = err
		return
	}

	characterInfo := resp.Result().(*api.CharacterInfoResponse)

	context.CharacterID = characterInfo.ID
	context.CharacterInfo = characterInfo
	m.Data = *characterInfo

	req := request.CharacterGetCurrencyAmount(api.Grynars)

	resp, err = helper.SendRequest(req)

	if err != nil {
		m.ErrMsg = err
		return
	}

	currencyResp := resp.Result().(*api.CurrencyResponse)

	m.CharacterVitalInfo.UpdateData(characterInfo, currencyResp.Amount)

	m.EquipmentSummary.UpdateData()
	m.EquipmentSummaryContainer.UpdateContent(m.EquipmentSummary)

	m.StatsSummary.UpdateData(characterInfo)
	m.StatsSummaryContainer.UpdateContent(m.StatsSummary)

	m.SkillSummary.UpdateData(characterInfo)
	m.SkillSummaryContainer.UpdateContent(m.SkillSummary)
}
