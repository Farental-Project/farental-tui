package maildetaileditor

import (
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/widgetfocusmanager"
	"farental/model"
	"farental/model/widgetmodel/list"
	"farental/style"

	teaList "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
)

type ListAttachmentModel struct {
	widgetfocusmanager.BaseFocusableWidget
	List *list.Model
	Data *[]teaList.Item
}

func NewListAttachment(width int, height int) *ListAttachmentModel {
	keymapContext := bubblehelp.NewKeymap(2)
	keymapContext.Style = style.MainHelpStyle
	keymapContext.NewKeyBinding(keybind.Up, true)
	keymapContext.NewKeyBinding(keybind.Down, true)
	keymapContext.NewKeyBinding(keybind.DKey, true)
	keymapContext.SetHelpDesc(keybind.DKey, lang.L("delete"))
	keymapContext.NewKeyBinding(keybind.AKey, true)
	keymapContext.SetHelpDesc(keybind.AKey, lang.L("add"))

	bubblehelp.RegisterContext(model.ContextMailDetailEditorAttachmentList, keymapContext)

	m := new(ListAttachmentModel)

	// Temporary data
	m.Data = &[]teaList.Item{
		ListItem{
			3, "Truite",
		},
		ListItem{
			5, "Fruit",
		},
		ListItem{
			3, "Toto",
		},
		ListItem{
			2, "Tissu",
		},
		ListItem{
			3, "Patate",
		},
	}

	m.List = list.New(*m.Data,
		ListItemDelegate{},
		width, height)
	m.List.SetShowHelp(false)
	m.List.SetShowTitle(false)
	m.List.SetShowFilter(false)
	m.List.SetShowStatusBar(false)
	m.List.SetShowPagination(false)

	return m
}

func (m *ListAttachmentModel) Init() tea.Cmd {
	return nil
}

func (m *ListAttachmentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.List.Update(msg)

	return m, nil
}

func (m *ListAttachmentModel) View() string {
	var containerStyle lipgloss.Style

	if m.Focused {
		containerStyle = style.FocusedStyle
	} else {
		containerStyle = style.BlurredStyle
	}

	return containerStyle.Render(m.List.View())
}

func (m *ListAttachmentModel) Focus() {
	m.BaseFocusableWidget.Focus()
	m.List.Focus()
	bubblehelp.SwitchContext(model.ContextMailDetailEditorAttachmentList)
}

func (m *ListAttachmentModel) Blur() {
	m.BaseFocusableWidget.Blur()
	m.List.Blur()
	bubblehelp.SwitchToPreviousContext()
}
