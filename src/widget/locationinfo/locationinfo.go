package locationinfo

import (
	"farental/art"
	"farental/core/data/api"
	"farental/internal/style"
	ftheme "farental/internal/theme"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
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
	w.title.Style = t.Style(ftheme.DimUnderlinedTextStyleID).
		AlignHorizontal(lipgloss.Center)
	w.title.SizeConstraint = true

	w.description = orvyn.NewSimpleRenderable("")
	w.description.Style = t.Style(theme.NormalTextStyleID).
		AlignHorizontal(lipgloss.Center)
	w.description.SizeConstraint = true

	w.layout = layout.NewMaxWidthVBoxLayout(0,
		w.title,
		w.description,
	)

	return w
}

func (w *Widget) Render() string {
	return w.GetStyle().Render(w.layout.Render())
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	w.layout.Resize(w.GetContentSize())
}

// GetMinSize measures an actual render. Render draws the title and the description
// at their natural height and ignores whatever height it is given, so declaring a
// flexible 10..12 made the layout reserve stretch rows this widget never draws in -
// rows the log and social panels could have used.
func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(30, lipgloss.Height(w.Render()))
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(45, lipgloss.Height(w.Render()))
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
	fmt.Fprintf(&b, "%s | %s",
		t.Style(theme.NeutralDimTextStyleID).Render(location.Type.Name),
		style.LocationBiomeStyle(location.Biome.Code).Render(location.Biome.Name))

	if len(location.Features) > 0 {
		b.WriteString("\n")

		for _, f := range location.Features {
			if features.Len() > 0 {
				fmt.Fprintf(&features, " %c ", art.CharBullet)
			}

			features.WriteString(f.Name)

		}

		b.WriteString(t.Style(theme.DimTextStyleID).Render(features.String()))
	}

	w.title.SetValue(b.String())
}
