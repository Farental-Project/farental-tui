package inventorystackinspect

import (
	"farental/core/data/api"
	ftheme "farental/internal/theme"
	"fmt"
	"strings"

	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
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
	t := orvyn.GetTheme()

	w.BaseWidget = orvyn.NewBaseWidget()

	titleStyle := t.Style(ftheme.DimUnderlinedTextStyleID)

	w.srName = orvyn.NewSimpleRenderable("")
	w.srName.Style = titleStyle
	w.srName.SizeConstraint = true
	w.srUnique = orvyn.NewSimpleRenderable(lokyn.L("Unique"))
	w.srUnique.Style = t.Style(theme.HighlightTextStyleID)
	w.srUnique.SetActive(false)
	w.srDescription = orvyn.NewSimpleRenderable("")
	w.srDescription.SizeConstraint = true

	w.srStatsTitle = orvyn.NewSimpleRenderable(
		fmt.Sprintf("\n%s", lokyn.L("Stats")))
	w.srStatsTitle.Style = titleStyle
	w.srStatsTitle.SizeConstraint = true
	w.srStats = orvyn.NewSimpleRenderable("")
	w.srStats.SizeConstraint = true
	w.srStatsTitle.SetActive(false)
	w.srStats.SetActive(false)

	w.srConditionsTitle = orvyn.NewSimpleRenderable(
		fmt.Sprintf("\n%s", lokyn.L("Equip conditions")))
	w.srConditionsTitle.Style = titleStyle
	w.srConditionsTitle.SizeConstraint = true
	w.srConditions = orvyn.NewSimpleRenderable("")
	w.srConditions.SizeConstraint = true
	w.srConditionsTitle.SetActive(false)
	w.srConditions.SetActive(false)

	w.srResultsTitle = orvyn.NewSimpleRenderable(
		fmt.Sprintf("\n%s", lokyn.L("Effects")))
	w.srResultsTitle.Style = titleStyle
	w.srResultsTitle.SizeConstraint = true
	w.srResults = orvyn.NewSimpleRenderable("")
	w.srResults.SizeConstraint = true
	w.srResultsTitle.SetActive(false)
	w.srResults.SetActive(false)

	w.layout = layout.NewMaxWidthVBoxLayout(1,
		w.srName,
		w.srUnique,
		w.srDescription,
		w.srStatsTitle,
		w.srStats,
		w.srConditionsTitle,
		w.srConditions,
		w.srResultsTitle,
		w.srResults,
	)

	return w
}

func (w *Widget) Render() string {
	return w.GetStyle().
		Height(w.GetContentSize().Height).
		Render(w.layout.Render())
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	w.layout.Resize(w.GetContentSize())
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

		if stack.Item.Conditions != nil {
			// Conditions
			for i, c := range *stack.Item.Conditions {
				if i > 0 {
					b.WriteString("\n")
				}

				b.WriteString(fmt.Sprintf("• %s", c))
			}
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
