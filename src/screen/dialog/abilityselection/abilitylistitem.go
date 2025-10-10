package abilityselection

import (
	"farental/core/data/api"
	ftheme "farental/internal/theme"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
)

type AbilityListItem struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	titleAbility *orvyn.SimpleRenderable
	titleSkill   *orvyn.SimpleRenderable
	description  *orvyn.SimpleRenderable

	powerLabel *orvyn.SimpleRenderable
	powerValue *orvyn.SimpleRenderable

	manaCostLabel *orvyn.SimpleRenderable
	manaCostValue *orvyn.SimpleRenderable

	cooldownLabel *orvyn.SimpleRenderable
	cooldownValue *orvyn.SimpleRenderable

	targetTypeLabel *orvyn.SimpleRenderable
	targetTypeValue *orvyn.SimpleRenderable

	layout *layout.VBoxFullLayout

	style lipgloss.Style

	data api.AbilityResponse

	contentSize orvyn.Size
}

func Constructor(data api.AbilityResponse) list.ListItem[api.AbilityResponse] {
	a := new(AbilityListItem)

	t := orvyn.GetTheme()

	a.BaseWidget = orvyn.NewBaseWidget()

	a.data = data

	a.titleAbility = orvyn.NewSimpleRenderable(data.Name)
	a.titleAbility.SizeConstraint = true
	a.titleAbility.Style = t.Style(ftheme.TitleUnderlinedTextStyleID)

	skillFmt := ""

	if data.SkillLevelMax > 0 {
		skillFmt = fmt.Sprintf("%s | %s : %d | %s : %d",
			data.SkillName,
			lokyn.L("Min"), data.SkillLevelMin,
			lokyn.L("Max"), data.SkillLevelMax)
	} else {
		skillFmt = fmt.Sprintf("%s | %s : %d",
			data.SkillName,
			lokyn.L("Min"), data.SkillLevelMin)
	}

	a.titleSkill = orvyn.NewSimpleRenderable(skillFmt)
	a.titleSkill.SizeConstraint = true
	a.titleSkill.Style = t.Style(ftheme.TitleUnderlinedTextStyleID).
		AlignHorizontal(lipgloss.Right)

	a.description = orvyn.NewSimpleRenderable(data.Description)
	a.description.SizeConstraint = true
	a.description.Style = t.Style(theme.DimTextStyleID).AlignVertical(lipgloss.Top)

	nsAlignRightStyle := t.Style(theme.NormalTextStyleID).AlignHorizontal(lipgloss.Right)
	dsAlignRightStyle := t.Style(theme.DimTextStyleID).AlignHorizontal(lipgloss.Right)

	a.powerLabel = orvyn.NewSimpleRenderable(fmt.Sprintf("%s :", lokyn.L("Power")))
	a.powerLabel.SizeConstraint = true
	a.powerLabel.Style = dsAlignRightStyle
	a.powerValue = orvyn.NewSimpleRenderable(fmt.Sprintf("%d", data.Power))
	a.powerValue.SizeConstraint = true
	a.powerValue.Style = nsAlignRightStyle

	a.manaCostLabel = orvyn.NewSimpleRenderable(fmt.Sprintf("%s :", lokyn.L("Mana cost")))
	a.manaCostLabel.SizeConstraint = true
	a.manaCostLabel.Style = dsAlignRightStyle
	a.manaCostValue = orvyn.NewSimpleRenderable(fmt.Sprintf("%d", data.ManaCost))
	a.manaCostValue.SizeConstraint = true
	a.manaCostValue.Style = nsAlignRightStyle

	a.cooldownLabel = orvyn.NewSimpleRenderable(fmt.Sprintf("%s :", lokyn.L("Cooldown")))
	a.cooldownLabel.SizeConstraint = true
	a.cooldownLabel.Style = dsAlignRightStyle
	a.cooldownValue = orvyn.NewSimpleRenderable(fmt.Sprintf("%d", data.Cooldown))
	a.cooldownValue.SizeConstraint = true
	a.cooldownValue.Style = nsAlignRightStyle

	a.targetTypeLabel = orvyn.NewSimpleRenderable(fmt.Sprintf("%s :", lokyn.L("Targeting")))
	a.targetTypeLabel.SizeConstraint = true
	a.targetTypeLabel.Style = dsAlignRightStyle

	targetType := ""

	if a.data.TargetGroup {
		targetType = lokyn.L("Group")
	} else {
		targetType = lokyn.L("Single")
	}

	a.targetTypeValue = orvyn.NewSimpleRenderable(targetType)
	a.targetTypeValue.SizeConstraint = true
	a.targetTypeValue.Style = nsAlignRightStyle

	titleLayout := layout.NewHBoxFixedRatioLayout(0, 0,
		1,
		[]layout.FixedRatioRenderable{
			layout.NewFixedRatioRenderable(0.5, a.titleAbility),
			layout.NewFixedRatioRenderable(0.5, a.titleSkill),
		})

	infoLayout := layout.NewMaxWidthVBoxLayout(0,
		[]orvyn.Renderable{
			layout.NewHBoxFixedRatioLayout(0, 0, 1,
				[]layout.FixedRatioRenderable{
					layout.NewFixedRatioRenderable(0.7, a.powerLabel),
					layout.NewFixedRatioRenderable(0.3, a.powerValue),
				}),
			layout.NewHBoxFixedRatioLayout(0, 0, 1,
				[]layout.FixedRatioRenderable{
					layout.NewFixedRatioRenderable(0.7, a.manaCostLabel),
					layout.NewFixedRatioRenderable(0.3, a.manaCostValue),
				}),
			layout.NewHBoxFixedRatioLayout(0, 0, 1,
				[]layout.FixedRatioRenderable{
					layout.NewFixedRatioRenderable(0.7, a.cooldownLabel),
					layout.NewFixedRatioRenderable(0.3, a.cooldownValue),
				}),
			layout.NewHBoxFixedRatioLayout(0, 0, 1,
				[]layout.FixedRatioRenderable{
					layout.NewFixedRatioRenderable(0.7, a.targetTypeLabel),
					layout.NewFixedRatioRenderable(0.3, a.targetTypeValue),
				}),
		})

	contentLayout := layout.NewHBoxFixedRatioLayout(0, 0, 1,
		[]layout.FixedRatioRenderable{
			layout.NewFixedRatioRenderable(0.6, a.description),
			layout.NewFixedRatioRenderable(0.4, infoLayout),
		})

	a.layout = layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(0, 0), 1,
		[]orvyn.Renderable{
			titleLayout,
			contentLayout,
		})

	a.OnBlur()

	return a
}

func (a *AbilityListItem) Resize(size orvyn.Size) {
	size.Height = 7

	a.BaseWidget.Resize(size)

	size.Width -= a.style.GetHorizontalFrameSize()
	size.Height -= a.style.GetVerticalFrameSize()

	a.contentSize = size
	a.layout.Resize(a.contentSize)
}

func (a *AbilityListItem) UpdateData(data api.AbilityResponse) {
	a.data = data
}

func (a *AbilityListItem) GetData() api.AbilityResponse {
	return a.data
}

func (a *AbilityListItem) Render() string {
	return a.style.
		Width(a.contentSize.Width).
		Height(a.contentSize.Height).
		Render(a.layout.Render())
}

func (a *AbilityListItem) OnFocus() {
	a.style = orvyn.GetTheme().Style(theme.FocusedWidgetStyleID)
}

func (a *AbilityListItem) OnBlur() {
	a.style = orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)
}

func (a *AbilityListItem) OnEnterInput() {

}

func (a *AbilityListItem) OnExitInput() {

}

func (a *AbilityListItem) FilterValue() string {
	var b strings.Builder

	b.WriteString(a.data.Name)
	b.WriteString(" ")
	b.WriteString(a.data.Description)
	b.WriteString(" ")
	b.WriteString(a.data.SkillName)

	return b.String()
}
