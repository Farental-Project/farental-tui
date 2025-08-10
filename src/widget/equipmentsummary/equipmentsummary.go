package equipmentsummary

import (
	"farental/core/data"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/style"
	"github.com/charmbracelet/lipgloss"
	"strings"
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
	contentSize    orvyn.Size
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.title = lang.L("Equipment")

	return w
}

func (w *Widget) Render() string {
	var col column
	var leftCol string
	var rightCol string

	size := w.contentSize

	if len(w.equipmentSlots) == 0 {
		return ""
	}

	col = column{}

	w.renderEquipmentSlot(data.SlotWeapon, true, &col)
	w.renderEquipmentSlot(data.SlotHead, true, &col)
	w.renderEquipmentSlot(data.SlotShoulder, true, &col)
	w.renderEquipmentSlot(data.SlotTorso, false, &col)

	leftCol = col.render()
	col.reset()

	w.renderEquipmentSlot(data.SlotShield, true, &col)
	w.renderEquipmentSlot(data.SlotHands, true, &col)
	w.renderEquipmentSlot(data.SlotLegs, true, &col)
	w.renderEquipmentSlot(data.SlotFeet, false, &col)

	rightCol = col.render()

	summary := lipgloss.JoinHorizontal(lipgloss.Top,
		style.TextStyle.Width(size.Width/2).Render(leftCol),
		style.TextStyle.Width(size.Width/2).Render(rightCol))

	content := lipgloss.JoinVertical(lipgloss.Left,
		style.DimUnderlinedTitleStyle.
			Width(size.Width).
			Render(w.title),
		summary)

	return style.BlurredStyle.
		Width(size.Width).Render(content)
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	size.Width -= style.BlurredStyle.GetHorizontalFrameSize()
	size.Height -= style.BlurredStyle.GetVerticalFrameSize()

	w.contentSize = size
}

func (w *Widget) renderEquipmentSlot(slotCode data.SlotCode, addReturn bool, column *column) {
	var itemNameStyle lipgloss.Style

	es := w.equipmentSlots[string(slotCode)]

	column.slotStr.WriteString(style.NormalStyle.Render(es.slot.Name))
	column.sepStr.WriteString(style.DimTextStyle.Render(" : "))

	if es.item.ID == 0 {
		itemNameStyle = style.NeutralDimTextStyle
	} else {
		itemNameStyle = style.DimTextStyle
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
	resp, err := helper.SendRequest(request.DataGetEquipmentSlots())

	if err != nil {
		return
	}

	slots = *resp.Result().(*[]api.BasicInfoResponse)

	// Get equipped items
	resp, err = helper.SendRequest(request.InventoryGetEquippedItems())

	if err != nil {
		return
	}

	equippedItems = *resp.Result().(*[]api.ItemResponse)

	for _, s := range slots {
		es := EquipmentSlot{
			slot: s,
			item: api.ItemResponse{
				ID:   0,
				Name: lang.L("None"),
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
