package characterselection

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/config"
	"farental/internal/context"
	"farental/internal/keybind"
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

	m.List = list.New(m.Items, ListItemDelegate{},
		style.LayoutWidth, 20)
	m.List.SetShowHelp(false)
	m.List.SetShowStatusBar(false)
	m.List.SetShowFilter(false)
	m.List.SetShowPagination(false)
	m.List.SetShowTitle(false)
	m.List.DisableQuitKeybindings()
	m.List.InfiniteScrolling = true

	m.Title = lang.L("Character selection")

	m.Keymap = config.ModularKeyMap{}

	m.Keymap.SetBindings([][]key.Binding{
		{
			keybind.Up,
			keybind.Down,
			keybind.Submit,
		},
		{
			keybind.Back,
			keybind.Quit,
			keybind.HelpClose,
		},
	})
	m.Keymap.SetEssentialBindings([]key.Binding{
		keybind.Back,
		keybind.Quit,
		keybind.HelpMore,
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
		m.initData()

		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return m, tea.Quit

		case key.Matches(msg, keybind.Submit):
			ok := m.submit()

			if ok {
				return context.ContentManager.SwitchContent(
					model.ContentGameDashboard)
			}

			return m, nil
		case key.Matches(msg, keybind.HelpMore):
			m.Help.ShowAll = !m.Help.ShowAll

			return m, nil

		case key.Matches(msg, keybind.Back):
			return context.ContentManager.SwitchContent(
				model.ContentLogin)
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
	tui.WriteString(style.ContainerStyle.Render(m.List.View()))
	tui.WriteString("\n\n")
	tui.WriteString(helpText)

	return lipgloss.Place(
		context.ContentManager.ScreenWidth,
		context.ContentManager.ScreenHeight,
		lipgloss.Center, lipgloss.Center,
		tui.String())
}

func (m *Model) initData() {
	m.loadCharacters()
	m.List.SetItems(m.Items)
}

func (m *Model) loadCharacters() {
	var characters *[]api.CharacterBasicResponse
	var ok bool

	req := request.CharacterGetAll()

	resp, err := req.Send()

	if err != nil {
		log.Println(err)
		return
	}

	characters, ok = resp.Result().(*[]api.CharacterBasicResponse)

	if !ok {
		log.Println("Invalid request response")
		return
	}

	m.Items = make([]list.Item, 0)

	for _, c := range *characters {
		item := ListItem{
			CharacterID:       c.ID,
			CharacterName:     c.FirstName + " " + c.LastName,
			CharacterRace:     c.RaceName,
			CharacterLocation: c.LocationName,
		}

		m.Items = append(m.Items, item)
	}
}

func (m *Model) submit() bool {
	selectedItem, ok := m.List.SelectedItem().(ListItem)

	if !ok {
		log.Println("Invalid item selected")
		return false
	}

	if selectedItem.CharacterID == 0 {
		log.Println("Selected character ID is 0")
		return false
	}

	req := request.CharacterSetActive(selectedItem.CharacterID)

	resp, err := req.Send()

	if err != nil {
		log.Println(err)
		return false
	}

	if resp.StatusCode() != 200 {
		log.Println(resp.Error())
		return false
	}

	context.CharacterID = selectedItem.CharacterID

	return true
}
