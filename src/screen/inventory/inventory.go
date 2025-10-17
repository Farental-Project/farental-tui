package inventory

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/widget/help"
	"farental/widget/inventorylistitem"
	"farental/widget/inventorystackinspect"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
)

type Screen struct {
	title         *orvyn.SimpleRenderable
	list          *list.Widget[api.StackResponse]
	inspector     *inventorystackinspect.Widget
	statusMessage *statusmessage.Widget
	help          *help.Widget

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable(lokyn.L("Inventory"))
	s.title.Style = t.Style(theme.TitleStyleID)

	s.list = list.New(inventorylistitem.Constructor)

	s.list.PreferredSize.Width = t.Size(ftheme.LayoutWidthSizeID) // border
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
				s.help,
			},
		),
	)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextInventory)

	s.loadInventory()
	s.list.FocusFirst()

	selectedItem := s.list.GetSelectedItem()

	s.updateInspector(&selectedItem)
	s.updateKeybind(&selectedItem.Item)

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch {
		case key.Matches(msg, keybind.Quit):
			return tea.Quit

		case key.Matches(msg, keybind.Esc):
			if s.list.FilterState() == list.Unfiltered {
				return orvyn.SwitchToPreviousScreen()
			}

		case key.Matches(msg, keybind.Help):
			if s.list.FilterState() != list.Filtering {
				bubblehelp.ShowAll = !bubblehelp.ShowAll
			}

			return nil
		}
	}

	cmd := s.list.Update(msg)

	selectedItem := s.list.GetSelectedItem()
	index := s.list.GetGlobalIndex()

	if selectedItem.ID > 0 {
		if s.inspector.GetCurrentStackItemID() != selectedItem.ItemID {
			s.updateInspector(&selectedItem)
			s.updateKeybind(&selectedItem.Item)
		}
	}

	if s.list.FilterState() == list.Filtering {
		return cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.UKey):
			if bubblehelp.IsKeybindVisible(keybind.UKey) {
				if s.checkRunningTask() {
					s.useItem(index, &selectedItem)
					return cmd
				}
			}

		case key.Matches(msg, keybind.EKey):
			if bubblehelp.IsKeybindVisible(keybind.EKey) {
				if s.checkRunningTask() {
					s.equipItem(&selectedItem)
					return cmd
				}
			}
		}
	}

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) loadInventory() {
	var inventory api.InventoryResponse

	resp, err := helper.SendRequest(request.InventoryGetFull())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	inventory = *resp.Result().(*api.InventoryResponse)

	s.list.SetItems(inventory.Stacks)
}

func (s *Screen) submit() bool {
	return false
}

func (s *Screen) updateInspector(item *api.StackResponse) {
	s.inspector.UpdateData(item)
}

func (s *Screen) useItem(index int, item *api.StackResponse) {
	req := request.InventoryUseItem(item.ItemID)

	_, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	item.Count--

	s.statusMessage.SetMessage(lokyn.L("Item used !"), statusmessage.SuccessMessage)

	if item.Count == 0 {
		s.list.RemoveItem(index)
		return
	}

	s.list.SetItem(index, *item)
}

func (s *Screen) equipItem(item *api.StackResponse) {
	req := request.InventoryEquipItem(item.ItemID)

	_, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	item.Count--

	s.statusMessage.SetMessage(lokyn.L("Item equipped !"), statusmessage.SuccessMessage)

	s.loadInventory()
}

func (s *Screen) updateKeybind(item *api.ItemResponse) {
	if item == nil {
		bubblehelp.SetKeybindVisible(keybind.UKey, false)
		bubblehelp.SetKeybindVisible(keybind.EKey, false)
		return
	}

	if item.IsUsable {
		bubblehelp.SetKeybindVisible(keybind.UKey, true)
	} else {
		bubblehelp.SetKeybindVisible(keybind.UKey, false)
	}

	if item.EquipmentSlot != nil {
		bubblehelp.SetKeybindVisible(keybind.EKey, true)
	} else {
		bubblehelp.SetKeybindVisible(keybind.EKey, false)
	}
}

func (s *Screen) checkRunningTask() bool {
	if context.RunningTask != nil {
		s.statusMessage.SetMessage(lokyn.L("A task is running. Claim it before doing this."), statusmessage.InformationMessage)
		return false
	}

	return true
}
