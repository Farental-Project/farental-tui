package inventory

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/internal/orvyn/layout"
	"farental/style"
	"farental/widget/filterablelist"
	"farental/widget/help"
	"farental/widget/inventorystackinspect"
	"farental/widget/statusmessage"
	"github.com/charmbracelet/bubbles/key"
	tealist "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
)

type Screen struct {
	orvyn.BaseScreen

	title         *orvyn.SimpleRenderable
	list          *filterablelist.Widget
	inspector     *inventorystackinspect.Widget
	statusMessage *statusmessage.Widget
	help          *help.Widget

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	s.title = orvyn.NewSimpleRenderable(lang.L("Inventory"))
	s.title.Style = style.TitleStyle

	s.list = filterablelist.New(ListItemDelegate{}, []tealist.Item{})

	s.list.PreferredSize.Width = style.LayoutWidth - 2 // border
	s.list.PreferredSize.Height = 80
	s.list.MinSize.Height = 13

	s.inspector = inventorystackinspect.New()
	s.statusMessage = statusmessage.New()

	s.help = help.New()

	inventoryLayout := layout.NewHBoxFixedRatioLayout(0, 1,
		0,
		[]layout.FixedRatioRenderable{
			layout.NewFixedRatioRenderable(0.60, s.list),
			layout.NewFixedRatioRenderable(0.40, s.inspector),
		},
	)

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(10, 4), 2,
			[]orvyn.Renderable{
				s.title,
				orvyn.VGap,
				inventoryLayout,
				s.statusMessage,
				orvyn.VGap,
				s.help,
			},
		),
	)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
	s.loadInventory()
	s.list.Select(0)
	s.updateInspector()

	return nil
}

func (s *Screen) OnExit() interface{} {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch {
		case key.Matches(msg, keybind.Quit):
			return tea.Quit

		case key.Matches(msg, keybind.Esc):
			if s.list.FilterState() == tealist.Unfiltered {
				return orvyn.SwitchToPreviousScreen()
			}

		case key.Matches(msg, keybind.Help):
			if s.list.FilterState() != tealist.Filtering {
				bubblehelp.ShowAll = !bubblehelp.ShowAll
			}

			return nil
		}
	}

	cmd := s.list.Update(msg)

	s.updateInspector()

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) loadInventory() {
	var inventory api.InventoryResponse
	var items []tealist.Item

	items = make([]tealist.Item, 0)

	resp, err := helper.SendRequest(request.InventoryGetFull())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	inventory = *resp.Result().(*api.InventoryResponse)

	for _, s := range inventory.Stacks {
		item := ListItem{
			Stack: s,
		}

		items = append(items, item)
	}

	s.list.SetItems(items)
}

func (s *Screen) submit() bool {
	return false
}

func (s *Screen) updateInspector() {
	item, ok := s.list.SelectedItem().(ListItem)

	if !ok {
		return
	}

	s.inspector.UpdateData(&item.Stack)
}
