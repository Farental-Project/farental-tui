package travelselection

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/config"
	"farental/internal/context"
	"farental/internal/lang"
	"farental/model"
	"farental/style"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"log"
	"strings"
)

type Model struct {
	List   list.Model
	Items  []list.Item
	Help   help.Model
	Keymap config.ModularKeyMap

	Title string
}

func New() Model {
	m := Model{}

	m.Help = help.New()

	m.Items = make([]list.Item, 0)

	m.List = list.New(m.Items,
		ListItemDelegate{},
		style.LayoutWidth, 45)
	m.List.SetShowHelp(false)

	m.Title = lang.L("Travel selection")

	m.Keymap = config.ModularKeyMap{}

	m.Keymap.SetBindings([][]key.Binding{
		{
			config.Up,
			config.Down,
			config.Submit,
		},
		{
			config.Help,
			config.Back,
			config.Quit,
		},
	})
	m.Keymap.SetEssentialBindings([]key.Binding{
		config.Help,
		config.Back,
		config.Quit,
	})

	return m
}

func (m Model) Init() tea.Cmd {
	return model.InitCmd
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case model.InitMsg:
		m.UpdateData()

		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, config.Quit):
			return m, tea.Quit

		case key.Matches(msg, config.Submit):
			ok := m.submit()

			if ok {
				return context.ContentManager.SwitchContent(
					model.ContentGameDashboard)
			}

			return m, nil
		case key.Matches(msg, config.Help):
			m.Help.ShowAll = !m.Help.ShowAll

			return m, nil

		case key.Matches(msg, config.Back):
			return context.ContentManager.SwitchContent(
				model.ContentGameDashboard)
		}
	}

	context.ContentManager.Update(msg)

	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	var tui strings.Builder

	helpText := m.Help.View(m.Keymap)

	title := style.TitleStyle.Render(m.Title)

	tui.WriteString(title)
	tui.WriteString("\n\n")
	tui.WriteString(style.ContainerStyle.Width(style.LayoutWidth).Render(m.List.View()))
	tui.WriteString("\n\n")
	tui.WriteString(helpText)

	return lipgloss.Place(
		context.ContentManager.ScreenWidth,
		context.ContentManager.ScreenHeight,
		lipgloss.Center, lipgloss.Center,
		tui.String())
}

func (m *Model) UpdateData() {
	m.loadTravels()
	m.List.SetItems(m.Items)
}

func (m *Model) loadTravels() {
	var travels []api.TravelResponse

	req := request.TravelGetAvailable()

	resp, err := req.Send()

	if err != nil {
		log.Println(err)
		return
	}

	travels = *resp.Result().(*[]api.TravelResponse)

	m.Items = make([]list.Item, 0)

	for _, t := range travels {
		item := ListItem{
			Travel: t,
		}

		m.Items = append(m.Items, item)
	}
}

func (m *Model) submit() bool {
	i, ok := m.List.SelectedItem().(ListItem)

	if !ok {
		return false
	}

	req := request.TravelStart(i.Travel.ID)

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
