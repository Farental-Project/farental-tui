package craftselection

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/model"
	"farental/model/widget/filterselectionlist"
	"farental/style"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"log"
	"strings"
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

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(model.InitCmd, m.FilterSelectionList.Init())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var mod tea.Model

	defer context.ContentManager.UpdateCurrentContent(m)

	switch msg := msg.(type) {
	case model.InitMsg:
		context.KeymapManager.SwitchContext(model.ContextFilterSelectionListIncDecPage)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Esc):
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

	return m, cmd
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(m.FilterSelectionList.ViewTitle())
	b.WriteString("\n\n")
	b.WriteString(m.FilterSelectionList.View())
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
