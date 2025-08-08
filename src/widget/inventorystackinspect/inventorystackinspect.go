package inventorystackinspect

import (
	"farental/core/data/api"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/internal/orvyn/layout"
	"farental/style"
	"fmt"
	"strings"
)

type Widget struct {
	orvyn.BaseWidget

	srName            *orvyn.SimpleRenderable
	srUnique          *orvyn.SimpleRenderable
	srDescription     *orvyn.SimpleRenderable
	srStatsTitle      *orvyn.SimpleRenderable
	srStats           *orvyn.SimpleRenderable
	srConditionsTitle *orvyn.SimpleRenderable
	srConditions      *orvyn.SimpleRenderable
	srResultsTitle    *orvyn.SimpleRenderable
	srResults         *orvyn.SimpleRenderable

	layout *layout.VBoxLayout

	currentStackItemID uint
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	titleStyle := style.DimBottomBorderStyle

	w.srName = orvyn.NewSimpleRenderable("")
	w.srName.Style = style.DimUnderlinedTitleStyle
	w.srName.SizeConstraint = true
	w.srUnique = orvyn.NewSimpleRenderable(lang.L("Unique"))
	w.srUnique.Style = style.SpecialHighlightStyle
	w.srUnique.SetActive(false)
	w.srDescription = orvyn.NewSimpleRenderable("")
	w.srDescription.SizeConstraint = true

	w.srStatsTitle = orvyn.NewSimpleRenderable(
		fmt.Sprintf("\n%s", lang.L("Stats")))
	w.srStatsTitle.Style = titleStyle
	w.srStatsTitle.SizeConstraint = true
	w.srStats = orvyn.NewSimpleRenderable("")
	w.srStats.SizeConstraint = true
	w.srStatsTitle.SetActive(false)
	w.srStats.SetActive(false)

	w.srConditionsTitle = orvyn.NewSimpleRenderable(
		fmt.Sprintf("\n%s", lang.L("Equip conditions")))
	w.srConditionsTitle.Style = titleStyle
	w.srConditionsTitle.SizeConstraint = true
	w.srConditions = orvyn.NewSimpleRenderable("")
	w.srConditions.SizeConstraint = true
	w.srConditionsTitle.SetActive(false)
	w.srConditions.SetActive(false)

	w.srResultsTitle = orvyn.NewSimpleRenderable(
		fmt.Sprintf("\n%s", lang.L("Effects")))
	w.srResultsTitle.Style = titleStyle
	w.srResultsTitle.SizeConstraint = true
	w.srResults = orvyn.NewSimpleRenderable("")
	w.srResults.SizeConstraint = true
	w.srResultsTitle.SetActive(false)
	w.srResults.SetActive(false)

	w.layout = layout.NewMaxWidthVBoxLayout(1,
		[]orvyn.Renderable{
			w.srName,
			w.srUnique,
			w.srDescription,
			w.srStatsTitle,
			w.srStats,
			w.srConditionsTitle,
			w.srConditions,
			w.srResultsTitle,
			w.srResults,
		},
	)

	return w
}

func (w *Widget) Render() string {
	return style.BlurredStyle.
		Height(w.layout.GetSize().Height).
		Render(w.layout.Render())
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	size.Width -= style.BlurredStyle.GetHorizontalFrameSize()
	size.Height -= style.BlurredStyle.GetVerticalFrameSize()

	w.layout.Resize(size)
}

func (w *Widget) UpdateData(stack *api.StackResponse) {
	var b strings.Builder
	var show bool

	w.currentStackItemID = stack.ItemID

	w.srName.SetValue(stack.Item.Name)
	w.srDescription.SetValue(stack.Item.Description)
	w.srUnique.SetActive(stack.Item.IsUnique)

	isEquipment := stack.Item.EquipmentSlot != nil

	if isEquipment {
		// Stats
		for i, s := range *stack.Item.EquipmentStats {
			if i > 0 {
				b.WriteString("\n")
			}

			b.WriteString(fmt.Sprintf("• %s : %d", s.Stat.Name, s.Value))
		}

		show = b.Len() > 0

		w.srStatsTitle.SetActive(show)
		w.srStats.SetActive(show)

		if show {
			w.srStats.SetValue(b.String())
			b.Reset()
		}

		// Conditions
		for i, c := range *stack.Item.Conditions {
			if i > 0 {
				b.WriteString("\n")
			}

			b.WriteString(fmt.Sprintf("• %s", c))
		}

		show = b.Len() > 0

		w.srConditionsTitle.SetActive(show)
		w.srConditions.SetActive(show)

		if show {
			w.srConditions.SetValue(b.String())
			b.Reset()
		}
	} else {
		w.srStatsTitle.SetActive(false)
		w.srStats.SetActive(false)
		w.srConditionsTitle.SetActive(false)
		w.srConditions.SetActive(false)
	}

	isUseable := stack.Item.Results != nil

	if isUseable {
		for i, r := range *stack.Item.Results {
			if i > 0 {
				b.WriteString("\n")
			}

			b.WriteString(fmt.Sprintf("• %s", r))
		}

		show = b.Len() > 0

		w.srResultsTitle.SetActive(show)
		w.srResults.SetActive(show)

		if show {
			w.srResults.SetValue(b.String())
		}
	} else {
		w.srResultsTitle.SetActive(false)
		w.srResults.SetActive(false)
	}
}

func (w *Widget) GetCurrentStackItemID() uint {
	return w.currentStackItemID
}
