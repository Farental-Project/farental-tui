package equipmentsummary

import (
	"farental/core/data"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	ftheme "farental/internal/theme"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
)

type EquipmentSlot struct {
	slot api.BasicInfoResponse
	item api.ItemResponse
}

type column struct {
	slotStr strings.Builder
	sepStr  strings.Builder
	itemStr strings.Builder
}

func (c *column) reset() {
	c.slotStr.Reset()
	c.sepStr.Reset()
	c.itemStr.Reset()
}

func (c *column) addReturn() {
	c.slotStr.WriteString("\n")
	c.sepStr.WriteString("\n")
	c.itemStr.WriteString("\n")
}

func (c *column) render() string {
	return lipgloss.JoinHorizontal(lipgloss.Center,
		c.slotStr.String(),
		c.sepStr.String(),
		c.itemStr.String())
}

type Widget struct {
	orvyn.BaseWidget

	title string

	equipmentSlots map[string]EquipmentSlot
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.title = lokyn.L("Equipment")

	return w
}

func (w *Widget) Render() string {
	var col column
	var leftCol string
	var rightCol string

	size := w.GetContentSize()

	t := orvyn.GetTheme()

	col = column{}

	w.renderEquipmentSlot(data.SlotWeapon, true, &col)
	w.renderEquipmentSlot(data.SlotHead, true, &col)
	w.renderEquipmentSlot(data.SlotShoulder, true, &col)
	w.renderEquipmentSlot(data.SlotTorso, false, &col)

	leftCol = col.render()
	col.reset()

	w.renderEquipmentSlot(data.SlotLeftHand, true, &col)
	w.renderEquipmentSlot(data.SlotHands, true, &col)
	w.renderEquipmentSlot(data.SlotLegs, true, &col)
	w.renderEquipmentSlot(data.SlotFeet, false, &col)

	rightCol = col.render()

	width1, width2 := orvyn.DivideSizeFull(size.Width)

	summary := lipgloss.JoinHorizontal(lipgloss.Top,
		t.Style(theme.NormalTextStyleID).Width(width1).Render(leftCol),
		t.Style(theme.NormalTextStyleID).Width(width2).Render(rightCol))

	content := lipgloss.JoinVertical(lipgloss.Left,
		t.Style(ftheme.DimUnderlinedTextStyleID).
			Width(size.Width).
			Render(w.title),
		summary)

	return w.GetStyle().
		Width(size.Width).Render(content)
}

func (w *Widget) renderEquipmentSlot(slotCode data.SlotCode, addReturn bool, column *column) {
	var itemNameStyle lipgloss.Style

	es := w.equipmentSlots[string(slotCode)]

	t := orvyn.GetTheme()

	column.slotStr.WriteString(t.Style(theme.NormalTextStyleID).Render(es.slot.Name))
	column.sepStr.WriteString(t.Style(theme.DimTextStyleID).Render(" : "))

	if es.item.ID == 0 {
		itemNameStyle = t.Style(theme.NeutralDimTextStyleID)
	} else {
		itemNameStyle = t.Style(theme.DimTextStyleID)
	}

	column.itemStr.WriteString(itemNameStyle.Render(es.item.Name))

	if addReturn {
		column.addReturn()
	}
}

func (w *Widget) UpdateData() {
	var slots []api.BasicInfoResponse
	var equippedItems []api.ItemResponse
	var Equipments map[string]EquipmentSlot

	Equipments = make(map[string]EquipmentSlot)

	// Get all slots
	slotsRes, err := helper.Fetch[[]api.BasicInfoResponse](request.DataGetEquipmentSlots())

	if err != nil {
		return
	}

	slots = *slotsRes

	// Get equipped items
	equippedItemsRes, err := helper.Fetch[[]api.ItemResponse](request.InventoryGetEquippedItems())

	if err != nil {
		return
	}

	equippedItems = *equippedItemsRes

	for _, s := range slots {
		es := EquipmentSlot{
			slot: s,
			item: api.ItemResponse{
				ID:   0,
				Name: lokyn.L("None"),
			},
		}

		for _, e := range equippedItems {
			if e.EquipmentSlot.Code == s.Code {
				es.item = e
			}
		}

		Equipments[s.Code] = es
	}

	w.equipmentSlots = Equipments
}
