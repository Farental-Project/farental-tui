package locationinfo

import (
	"farental/art"
	"farental/core/data/api"
	"farental/layout"
	"farental/style"
	"farental/widget/label"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"sort"
	"strings"
)

type Widget struct {
	orvyn.BaseWidget

	title       *label.Widget
	description *label.Widget

	layout *layout.VBoxLayout
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.title = label.New("")
	w.title.Style = style.DimBottomBorderStyle.
		AlignHorizontal(lipgloss.Center)

	w.description = label.New("")
	w.description.Style = style.NormalStyle.
		AlignHorizontal(lipgloss.Center)

	w.layout = layout.NewMaxWidthVBoxLayout(0,
		[]orvyn.Renderable{
			w.title,
			w.description,
		})

	return w
}

func (w *Widget) Render() string {
	return style.BlurredStyle.Render(w.layout.Render())
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	size.Width -= style.BlurredStyle.GetHorizontalFrameSize()
	size.Height -= style.BlurredStyle.GetVerticalFrameSize()

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

	b.WriteString(style.BoldTextStyle.Render(location.Name))
	b.WriteString("\n")
	b.WriteString(style.NeutralDimTextStyle.Render(location.Continent.Name))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%s | %s",
		style.NeutralDimTextStyle.Render(location.Type.Name),
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

		b.WriteString(style.DimTextStyle.Render(features.String()))
	}

	w.title.SetValue(b.String())
}
