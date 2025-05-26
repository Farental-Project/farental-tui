package characterselection

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal"
	"farental/internal/lang"
	"farental/model"
	"farental/style"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"log"
	"strings"
)

type Model struct {
	ctx *internal.AppCtx

	List  list.Model
	Items []list.Item
}

func New(ctx *internal.AppCtx) Model {
	m := Model{ctx: ctx}

	m.Items = make([]list.Item, 0)

	m.List = list.New(m.Items, ListItemDelegate{}, 30, 20)
	m.List.SetShowHelp(false)
	m.List.SetShowStatusBar(false)
	m.List.SetShowFilter(false)
	m.List.SetShowPagination(false)
	m.List.Title = lang.L("Character selection")
	m.List.Styles.Title = style.TitleStyle
	m.List.DisableQuitKeybindings()
	m.List.InfiniteScrolling = true

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
		s := msg.String()
		switch s {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	m.ctx.ContentManager.Update(msg)

	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(m.List.View())

	tui := style.ContainerStyle.Render(b.String())

	return lipgloss.Place(
		m.ctx.ContentManager.ScreenWidth, m.ctx.ContentManager.ScreenHeight,
		lipgloss.Center, lipgloss.Center,
		tui)
}

func (m *Model) initData() {
	m.loadCharacters()
	m.List.SetItems(m.Items)
}

func (m *Model) loadCharacters() {
	var characters *[]api.CharacterBasicInfoResponse
	var ok bool

	req := request.CharacterGetAll()

	resp, err := req.Send()

	if err != nil {
		log.Println(err)
		return
	}

	characters, ok = resp.Result().(*[]api.CharacterBasicInfoResponse)

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
