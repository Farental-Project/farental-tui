package equipmentsummary

import (
	"farental/core/data"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/lang"
	"farental/style"
	tea "github.com/charmbracelet/bubbletea"
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

type Model struct {
	equipmentSlots map[string]EquipmentSlot
	width          int
}

func New(width int) Model {
	m := Model{
		width: width,
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	var col column
	var leftCol string
	var rightCol string

	if len(m.equipmentSlots) == 0 {
		return ""
	}

	col = column{}

	m.renderEquipmentSlot(data.SlotWeapon, true, &col)
	m.renderEquipmentSlot(data.SlotHead, true, &col)
	m.renderEquipmentSlot(data.SlotShoulder, true, &col)
	m.renderEquipmentSlot(data.SlotTorso, false, &col)

	leftCol = col.render()
	col.reset()

	m.renderEquipmentSlot(data.SlotShield, true, &col)
	m.renderEquipmentSlot(data.SlotHands, true, &col)
	m.renderEquipmentSlot(data.SlotLegs, true, &col)
	m.renderEquipmentSlot(data.SlotFeet, false, &col)

	rightCol = col.render()

	return lipgloss.JoinHorizontal(lipgloss.Top,
		style.TextStyle.Width(m.width/2).Render(leftCol),
		style.TextStyle.Width(m.width/2).Render(rightCol))
}

func (m Model) renderEquipmentSlot(slotCode data.SlotCode, addReturn bool, column *column) {
	es := m.equipmentSlots[string(slotCode)]

	column.slotStr.WriteString(style.TitleStyle.Render(es.slot.Name))
	column.sepStr.WriteString(style.TitleStyle.Render(" : "))
	column.itemStr.WriteString(style.DimTextStyle.Render(es.item.Name))

	if addReturn {
		column.addReturn()
	}
}

func (m *Model) UpdateData() {
	var slots []api.BasicInfoResponse
	var equippedItems []api.ItemResponse
	var Equipments map[string]EquipmentSlot

	Equipments = make(map[string]EquipmentSlot)

	// Get all slots
	req := request.DataGetEquipmentSlots()

	resp, err := req.Send()

	if err != nil {
		return
	}

	err = helper.ExtractError(resp)

	if err != nil {
		return
	}

	slots = *resp.Result().(*[]api.BasicInfoResponse)

	// Get equipped items
	req = request.InventoryGetEquippedItems()

	resp, err = req.Send()

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

	m.equipmentSlots = Equipments
}
