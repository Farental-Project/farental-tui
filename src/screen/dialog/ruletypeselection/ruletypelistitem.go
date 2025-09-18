package ruletypeselection

import (
	"farental/core/data/api"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
	"strings"
)

type RuleTypeListItem struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	style lipgloss.Style

	data *api.ScriptRuleTypeResponse

	contentSize orvyn.Size
}

func Constructor(data *api.ScriptRuleTypeResponse) list.IListItem {
	w := new(RuleTypeListItem)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.data = data

	w.OnBlur()

	return w
}

func (w *RuleTypeListItem) Resize(size orvyn.Size) {
	size.Height = 3

	w.BaseWidget.Resize(size)

	size.Width -= w.style.GetHorizontalFrameSize()
	size.Height -= w.style.GetVerticalFrameSize()

	w.contentSize = size
}

func (r *RuleTypeListItem) Render() string {
	t := orvyn.GetTheme()

	str := r.style.Width(r.contentSize.Width).
		Height(r.contentSize.Height).Render(
		fmt.Sprintf("%s\n%s",
			t.Style(theme.HighlightTextStyleID).Render(r.data.Name),
			t.Style(theme.DimTextStyleID).
				Width(r.contentSize.Width).Render(r.data.Description),
		),
	)

	return str
}

func (r *RuleTypeListItem) OnFocus() {
	r.style = orvyn.GetTheme().Style(theme.FocusedWidgetStyleID)
}

func (r *RuleTypeListItem) OnBlur() {
	r.style = orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)
}

func (r *RuleTypeListItem) OnEnterInput() {
}

func (r *RuleTypeListItem) OnExitInput() {
}

func (r *RuleTypeListItem) FilterValue() string {
	var b strings.Builder

	b.WriteString(r.data.Name)
	b.WriteString(" ")
	b.WriteString(r.data.Description)

	return b.String()
}
