package mailbox

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
	"strings"
)

type Model struct {
	FilterSelectionList filterselectionlist.Model
}

func New() Model {
	m := Model{}

	m.FilterSelectionList = filterselectionlist.New(
		lang.L("Mailbox"),
		ListItemDelegate{},
		m.loadData,
		m.submit)

	m.FilterSelectionList.CustomEnterDesc = lang.L("open mail")

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
		context.KeymapManager.SwitchContext(model.ContextFilterSelectionListBasic)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Esc):
			if m.FilterSelectionList.List.FilterState() == list.Unfiltered {
				return context.ContentManager.
					SwitchContent(m, model.ContentGameDashboard)
			}
		}
	case model.SwitchContentMsg:
		// TODO: goto mail detail screen (How to pass value to new screen ?)
		return context.ContentManager.SwitchContent(m, model.ContentGameDashboard)
	}

	mod, cmd = m.FilterSelectionList.Update(msg)

	modFSL, ok := mod.(filterselectionlist.Model)

	if !ok {
		return mod, cmd
	}

	m.FilterSelectionList = modFSL

	context.ContentManager.Update(msg)

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
	var mails []api.MailBasicResponse
	var items []list.Item

	items = make([]list.Item, 0)

	resp, err := helper.SendRequest(request.MailGetAll())

	if err != nil {
		fsl.ErrMsg = err
		return items
	}

	mails = *resp.Result().(*[]api.MailBasicResponse)

	for _, a := range mails {
		item := NewListItem(a)

		items = append(items, item)
	}

	return items
}

func (m *Model) submit(fsl *filterselectionlist.Model) bool {
	// TODO: Switch to mail detail screen
	// For this I need to rework the filter list widget.
	// I cannot let the component manage the next screen part.
	// I guess it's all right with the msg I've put into place.

	return false
}
