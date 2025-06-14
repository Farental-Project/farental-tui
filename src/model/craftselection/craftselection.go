package craftselection

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/lang"
	"farental/model/widget/filterselectionlist"
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
		lang.L("Craft selection"),
		ListItemDelegate{},
		m.loadData,
		m.submit)

	m.FilterSelectionList.SetShowExtraKeybinds(true, true)

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

	return m, cmd
}

func (m Model) View() string {
	return m.FilterSelectionList.View()
}

func (m *Model) loadData(fsl *filterselectionlist.Model) []list.Item {
	var crafts []api.RecipeResponse
	var items []list.Item

	items = make([]list.Item, 0)

	req := request.CraftGetAvailable()

	resp, err := req.Send()

	if err != nil {
		fsl.ErrMsg = helper.ConnectionError()
		return items
	}

	fsl.ErrMsg = helper.ExtractError(resp)

	if fsl.ErrMsg != nil {
		return items
	}

	crafts = *resp.Result().(*[]api.RecipeResponse)

	for _, c := range crafts {
		item := NewListItem(&c)

		items = append(items, item)
	}

	return items
}

func (m *Model) submit(fsl *filterselectionlist.Model) bool {
	i, ok := fsl.List.SelectedItem().(ListItem)

	if !ok {
		return false
	}

	req := request.CraftStart(i.CraftRecipe.ID, i.Amount)

	resp, err := req.Send()

	if err != nil {
		fsl.ErrMsg = helper.ConnectionError()
		return false
	}

	fsl.ErrMsg = helper.ExtractError(resp)

	if fsl.ErrMsg != nil {
		return false
	}

	if resp.StatusCode() != 200 {
		log.Println(resp.Error())
		return false
	}

	return true
}
