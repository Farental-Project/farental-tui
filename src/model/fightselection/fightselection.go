package fightselection

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/lang"
	"farental/model/filterselectionlist"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"log"
)

type Model struct {
	FilterSelectionList filterselectionlist.Model
}

func New() Model {
	m := Model{}

	m.FilterSelectionList = filterselectionlist.New(
		lang.L("Fight selection"),
		ListItemDelegate{},
		m.loadData,
		m.submit)

	m.FilterSelectionList.ShowIncreaseDecrease = true

	return m
}

func (m Model) Init() tea.Cmd {
	return m.FilterSelectionList.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var mod tea.Model

	mod, cmd = m.FilterSelectionList.Update(msg)

	modFSL, ok := mod.(filterselectionlist.Model)

	if !ok {
		return mod, cmd
	}

	m.FilterSelectionList = modFSL

	return m, cmd
}

func (m Model) View() string {
	return m.FilterSelectionList.View()
}

func (m *Model) loadData() []list.Item {
	var fights []api.FightCompositionResponse
	var items []list.Item

	items = make([]list.Item, 0)

	req := request.FightGetAvailable()

	resp, err := req.Send()

	if err != nil {
		log.Println(err)
		return items
	}

	fights = *resp.Result().(*[]api.FightCompositionResponse)

	for _, f := range fights {
		item := NewListItem(f)

		items = append(items, item)
	}

	return items
}

func (m *Model) submit(fsl *filterselectionlist.Model) bool {
	i, ok := fsl.List.SelectedItem().(ListItem)

	if !ok {
		return false
	}

	req := request.FightStart(i.FightCompo.ID)

	resp, err := req.Send()

	if err != nil {
		log.Println(err)
		return false
	}

	if resp.StatusCode() != 200 {
		log.Println(resp.Error())
		return false
	}

	return true
}
