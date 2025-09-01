package locationinfo

import (
	"farental/art"
	"farental/core/data/api"
	"farental/internal/style"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"sort"
	"strings"
)

type Widget struct {
	orvyn.BaseWidget

	title       *orvyn.SimpleRenderable
	description *orvyn.SimpleRenderable

	layout *layout.VBoxLayout
}

func New() *Widget {
	w := new(Widget)

	t := orvyn.GetTheme()

	w.BaseWidget = orvyn.NewBaseWidget()

	w.title = orvyn.NewSimpleRenderable("")
	w.title.Style = lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Center)

	w.description = orvyn.NewSimpleRenderable("")
	w.description.Style = t.Style(theme.NormalTextStyleID).
		AlignHorizontal(lipgloss.Center)

	w.layout = layout.NewMaxWidthVBoxLayout(0,
		[]orvyn.Renderable{
			w.title,
			w.description,
		})

	return w
}

func (w *Widget) Render() string {
	return orvyn.GetTheme().Style(theme.BlurredWidgetStyleID).
		Render(w.layout.Render())
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	st := orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)

	size.Width -= st.GetHorizontalFrameSize()
	size.Height -= st.GetVerticalFrameSize()

	w.layout.Resize(size)
}

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(30, 10)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(45, 12)
}

func (w *Widget) UpdateData(location *api.LocationResponse) {
	w.constructTitle(location)

	w.description.SetValue(location.Description)
}

func (w *Widget) constructTitle(location *api.LocationResponse) {
	var b strings.Builder
	var features strings.Builder

	t := orvyn.GetTheme()

	b.WriteString(t.Style(theme.NeutralTextStyleID).
		Bold(true).Render(location.Name))
	b.WriteString("\n")
	b.WriteString(t.Style(theme.NeutralDimTextStyleID).Render(location.Continent.Name))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%s | %s",
		t.Style(theme.NeutralDimTextStyleID).Render(location.Type.Name),
		style.LocationBiomeStyle(location.Biome.Code).Render(location.Biome.Name)),
	)

	if len(location.Features) > 0 {
		b.WriteString("\n")

		sort.Slice(location.Features, func(i, j int) bool {
			return location.Features[i].Name < location.Features[j].Name
		})

		for _, f := range location.Features {
			if !f.IsAction {
				continue
			}

			if features.Len() > 0 {
				features.WriteString(fmt.Sprintf(" %c ", art.CharBullet))
			}

			features.WriteString(f.Name)

		}

		b.WriteString(t.Style(theme.DimTextStyleID).Render(features.String()))
	}

	w.title.SetValue(b.String())
}
