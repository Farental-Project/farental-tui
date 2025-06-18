package inventory

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/lang"
	"farental/model/widget/filterselectionlist"
	"farental/model/widget/itemdetail"
	"farental/style"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	FilterSelectionList filterselectionlist.Model
	ItemDetail          itemdetail.Model
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

	m.ItemDetail = itemdetail.New(nil)

	return m
}

func (m Model) Init() tea.Cmd {
	return m.FilterSelectionList.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var mod tea.Model

	defer context.ContentManager.UpdateCurrentContent(m)

	mod, cmd = m.FilterSelectionList.Update(msg)

	modFSL, ok := mod.(filterselectionlist.Model)

	if !ok {
		return mod, cmd
	}

	m.FilterSelectionList = modFSL

	selectedItem := m.FilterSelectionList.List.SelectedItem().(ListItem)

	m.ItemDetail.UpdateData(&selectedItem.Stack)

	return m, cmd
}

func (m Model) View() string {
	itemDetail := style.ContainerStyle.Width(35).
		Height(m.FilterSelectionList.List.Height()).
		Render(m.ItemDetail.View())

	return lipgloss.JoinHorizontal(lipgloss.Left,
		itemDetail,
		m.FilterSelectionList.View(),
	)
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
