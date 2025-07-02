package inventory

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/model"
	"farental/model/widget/filterselectionlist"
	"farental/model/widget/itemdetail"
	"farental/style"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type Model struct {
	FilterSelectionList filterselectionlist.Model
	ItemDetail          itemdetail.Model
	SuccessMsg          string
}

func New() Model {
	m := Model{}

	m.FilterSelectionList = filterselectionlist.New(
		lang.L("Inventory"),
		ListItemDelegate{},
		m.loadData,
		m.submit)

	m.FilterSelectionList.Width = 32
	m.FilterSelectionList.List.SetWidth(32)
	m.FilterSelectionList.List.SetHeight(20)

	m.ItemDetail = itemdetail.New(35)

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(model.InitCmd)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var mod tea.Model

	defer context.ContentManager.UpdateCurrentContent(m)

	switch msgType := msg.(type) {
	case model.InitMsg:
		context.KeymapManager.SwitchContext(model.ContextInventory)

		m.FilterSelectionList.UpdateData()
		m.ItemDetail.UpdateData(nil)

		m.updateKeybind(nil)

	case tea.KeyMsg:
		switch {
		case key.Matches(msgType, keybind.Esc):
			if m.FilterSelectionList.List.FilterState() == list.Unfiltered {
				return context.ContentManager.
					SwitchContent(m, model.ContentGameDashboard)
			}
		}
	}

	mod, cmd = m.FilterSelectionList.Update(msg)

	modFSL, ok := mod.(filterselectionlist.Model)

	if !ok {
		return mod, cmd
	}

	m.FilterSelectionList = modFSL

	selectedIndex := m.FilterSelectionList.List.GlobalIndex()
	selectedItem, ok := m.FilterSelectionList.
		List.SelectedItem().(ListItem)

	if ok {
		if selectedItem.Stack.ItemID != m.ItemDetail.GetDataItemID() {
			m.ItemDetail.UpdateData(&selectedItem.Stack)
			m.updateKeybind(&selectedItem.Stack.Item)
		}
	} else {
		return m, cmd
	}

	switch msgType := msg.(type) {
	case tea.KeyMsg:
		m.SuccessMsg = ""
		if m.FilterSelectionList.List.FilterState() == list.Filtering {
			return m, cmd
		}

		switch {
		case key.Matches(msgType, keybind.Use):
			if selectedItem.Stack.Item.IsUsable {
				m.useItem(selectedItem, selectedIndex)
			}
		case key.Matches(msgType, keybind.Equip):
			if selectedItem.Stack.Item.EquipmentSlot != nil {
				m.equipItem(selectedItem, selectedIndex)
			}
		}
	}

	return m, cmd
}

func (m Model) View() string {
	var b strings.Builder

	itemDetail := style.ContainerStyle.Width(m.ItemDetail.GetWidth()).
		Render(m.ItemDetail.View())

	b.WriteString(m.FilterSelectionList.ViewTitle())
	b.WriteString("\n\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		m.FilterSelectionList.View(),
		itemDetail,
	))

	if len(m.SuccessMsg) > 0 {
		b.WriteString("\n")
		b.WriteString(style.TitleStyle.Render(m.SuccessMsg))
	}

	b.WriteString("\n")
	b.WriteString(m.FilterSelectionList.ViewError())
	b.WriteString("\n")
	b.WriteString(context.KeymapManager.View(style.LayoutWidth))

	return lipgloss.Place(
		context.ContentManager.ScreenWidth,
		context.ContentManager.ScreenHeight,
		lipgloss.Center, lipgloss.Center,
		b.String())

}

func (m *Model) loadData(fsl *filterselectionlist.Model) []list.Item {
	var inventory api.InventoryResponse
	var items []list.Item

	items = make([]list.Item, 0)

	req := request.InventoryGetFull()

	resp, err := req.Send()

	if err != nil {
		fsl.ErrMsg = helper.ConnectionError()
		return items
	}

	fsl.ErrMsg = helper.ExtractError(resp)

	if fsl.ErrMsg != nil {
		return items
	}

	inventory = *resp.Result().(*api.InventoryResponse)

	for _, s := range inventory.Stacks {
		item := ListItem{
			Stack: s,
		}

		items = append(items, item)
	}

	return items
}

func (m *Model) submit(fsl *filterselectionlist.Model) bool {
	return false
}

func (m *Model) useItem(selectedItem ListItem, index int) {
	req := request.InventoryUseItem(selectedItem.Stack.ItemID)

	resp, err := req.Send()

	if err != nil {
		m.FilterSelectionList.ErrMsg = helper.ConnectionError()
		return
	}

	m.FilterSelectionList.ErrMsg = helper.ExtractError(resp)

	if m.FilterSelectionList.ErrMsg != nil {
		return
	}

	selectedItem.Stack.Count--

	m.SuccessMsg = lang.L("Item used !")

	if selectedItem.Stack.Count == 0 {
		m.FilterSelectionList.List.RemoveItem(index)
		return
	}

	m.FilterSelectionList.List.SetItem(index, selectedItem)
}

func (m *Model) equipItem(selectedItem ListItem, index int) {
	req := request.InventoryEquipItem(selectedItem.Stack.ItemID)

	resp, err := req.Send()

	if err != nil {
		m.FilterSelectionList.ErrMsg = helper.ConnectionError()
		return
	}

	m.FilterSelectionList.ErrMsg = helper.ExtractError(resp)

	if m.FilterSelectionList.ErrMsg != nil {
		return
	}

	selectedItem.Stack.Count--

	m.SuccessMsg = lang.L("Item equipped !")

	m.FilterSelectionList.UpdateData()
}

func (m *Model) updateKeybind(item *api.ItemResponse) {
	if item == nil {
		context.KeymapManager.SetKeybindVisible(keybind.Use, false)
		context.KeymapManager.SetKeybindVisible(keybind.Equip, false)
		return
	}

	if item.IsUsable {
		context.KeymapManager.SetKeybindVisible(keybind.Use, true)
	} else {
		context.KeymapManager.SetKeybindVisible(keybind.Use, false)
	}

	if item.EquipmentSlot != nil {
		context.KeymapManager.SetKeybindVisible(keybind.Equip, true)
	} else {
		context.KeymapManager.SetKeybindVisible(keybind.Equip, false)
	}
}
