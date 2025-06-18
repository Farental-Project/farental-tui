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
	m.FilterSelectionList.List.SetHeight(25)

	m.ItemDetail = itemdetail.New(35)

	return m
}

func (m Model) Init() tea.Cmd {
	return m.FilterSelectionList.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var mod tea.Model

	defer context.ContentManager.UpdateCurrentContent(m)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Back):
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

	selectedItem, ok := m.FilterSelectionList.
		List.SelectedItem().(ListItem)

	if ok {
		if selectedItem.Stack.ItemID != m.ItemDetail.GetDataItemID() {
			m.ItemDetail.UpdateData(&selectedItem.Stack)
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
	b.WriteString("\n")
	b.WriteString(m.FilterSelectionList.ViewError())
	b.WriteString("\n")
	b.WriteString(m.FilterSelectionList.ViewHelp())

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
