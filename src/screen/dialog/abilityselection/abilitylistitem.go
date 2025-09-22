package abilityselection

import (
	"farental/core/data/api"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
)

type AbilityListItem struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	style lipgloss.Style

	data *api.AbilityResponse

	contentSize orvyn.Size
}

func Constructor(data *api.AbilityResponse) list.IListItem {
	w := new(AbilityListItem)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.data = data

	w.OnBlur()

	return w
}

func (a *AbilityListItem) Resize(size orvyn.Size) {
	size.Height = 7

	a.BaseWidget.Resize(size)

	size.Width -= a.style.GetHorizontalFrameSize()
	size.Height -= a.style.GetVerticalFrameSize()

	a.contentSize = size
}

func (a *AbilityListItem) Render() string {
	return ""
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
