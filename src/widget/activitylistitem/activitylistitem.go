package activitylistitem

import (
	"farental/core/data/api"
	"farental/internal/helper"
	"farental/widget/multivalueselector"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
	"strings"
)

type DurationData struct {
	api.DurationResponse
}

func (d DurationData) RenderValue() string {
	return helper.HoursDecFormat(d.Duration)
}

type Data struct {
	api.ActivityResponse
	DurationIndex int
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	name           *orvyn.SimpleRenderable
	description    *orvyn.SimpleRenderable
	skillName      *orvyn.SimpleRenderable
	durationSelect *multivalueselector.Widget[DurationData]

	style lipgloss.Style

	data *Data

	contentSize orvyn.Size

	layout *layout.HBoxFixedRatio
}

func Constructor(data *Data) list.IListItem {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.data = data

	w.OnBlur()

	t := orvyn.GetTheme()

	w.name = orvyn.NewSimpleRenderable(data.Name)
	w.name.Style = t.Style(theme.TitleStyleID)
	w.description = orvyn.NewSimpleRenderable(data.Description)
	w.description.Style = t.Style(theme.NormalTextStyleID)
	w.skillName = orvyn.NewSimpleRenderable(data.Skill.Name)
	w.skillName.Style = t.Style(theme.DimSecondaryTextStyleID)

	w.durationSelect = multivalueselector.New[DurationData]()

	leftLayout := layout.NewMaxWidthVBoxFullLayout(
		orvyn.NewSize(0, 0), 1,
		[]orvyn.Renderable{
			w.name,
			w.description,
		},
	)

	// TODO: Will probably need to manage alignment customization
	rightLayout := layout.NewMaxWidthVBoxLayout(
		0,
		[]orvyn.Renderable{
			w.skillName,
		},
	)

	w.layout = layout.NewHBoxFixedRatioLayout(
		0, 1, 0,
		[]layout.FixedRatioRenderable{
			layout.NewFixedRatioRenderable(
				0.8, leftLayout),
			layout.NewFixedRatioRenderable(
				0.2, rightLayout),
		},
	)

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	size.Width -= w.style.GetHorizontalFrameSize()
	size.Height -= w.style.GetVerticalFrameSize()

	w.layout.Resize(size)
	w.contentSize = size
}

func (w *Widget) Render() string {
	return w.style.
		Width(w.contentSize.Width).
		Height(w.contentSize.Height).
		Render(w.layout.Render())
}

func (w *Widget) OnFocus() {
	w.style = orvyn.GetTheme().Style(theme.FocusedWidgetStyleID)
}

func (w *Widget) OnBlur() {
	w.style = orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)
}

func (w *Widget) OnEnterInput() {}

func (w *Widget) OnExitInput() {}

func (w *Widget) FilterValue() string {
	var b strings.Builder

	b.WriteString(w.data.Name)
	b.WriteString(" ")
	b.WriteString(w.data.Description)
	b.WriteString(" ")
	b.WriteString(w.data.Skill.Name)

	return b.String()
}

func (w *Widget) loadDurations() {
	durations := w.data.Duration.Durations

	durationValues := make(map[string]DurationData)
	keys := make([]string, len(durations))

	for i, v := range durations {
		keys[i] = fmt.Sprintf("%f", v.Duration)
		durationValues[keys[i]] = DurationData{
			DurationResponse: v,
		}
	}

	w.durationSelect.SetValues(keys, durationValues)
	w.durationSelect.SetSelected(0)
}
