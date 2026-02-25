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
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type inventoryMode uint8

const (
	modeInventory inventoryMode = iota
	modeEquipped
)

type Screen struct {
	mode inventoryMode

	inventoryTitle string
	equippedTitle  string

	title           *orvyn.SimpleRenderable
	stackCountTitle *orvyn.SimpleRenderable
	list            *widgetlist.Widget[api.StackResponse]
	inspector       *inventorystackinspect.Widget
	statusMessage   *statusmessage.Widget
	help            *help.Widget

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.inventoryTitle = lokyn.L("Inventory")
	s.equippedTitle = lokyn.L("Equipped items")

	s.title = orvyn.NewSimpleRenderable(s.inventoryTitle)
	s.title.Style = t.Style(theme.TitleStyleID)

	s.stackCountTitle = orvyn.NewSimpleRenderable("")
	s.stackCountTitle.Style = t.Style(theme.NormalTextStyleID)

	s.list = widgetlist.New(inventorylistitem.Constructor)

	s.list.SetPreferredSize(orvyn.NewSize(t.Size(ftheme.LayoutWidthSizeID), 80))
	s.list.SetMinSize(orvyn.NewSize(30, 13))

	s.inspector = inventorystackinspect.New()
	s.statusMessage = statusmessage.New()

	s.help = help.New()

	s.mode = modeInventory

	inventoryElements := []layout.FixedRatioRenderable{
		layout.NewFixedRatioRenderable(0.60, s.list),
		layout.NewFixedRatioRenderable(0.40, s.inspector),
	}

	inventoryLayout := layout.NewHBoxFixedRatioLayout(0, 1, 0, inventoryElements...)

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(10, 4), 3,
			s.title,
			s.stackCountTitle,
			orvyn.VGap,
			inventoryLayout,
			s.statusMessage,
			s.help,
		),
	)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextInventory)

	s.statusMessage.Reset()

	s.loadInventory()
	s.list.FocusFirst()

	selectedItem := s.list.GetSelectedItem()

	s.mode = modeInventory

	s.title.SetValue(s.inventoryTitle)

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
			if s.list.FilterState() == widgetlist.Unfiltered {
				return orvyn.SwitchToPreviousScreen()
			}

		case key.Matches(msg, keybind.Help):
			if s.list.FilterState() != widgetlist.Filtering {
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

	if s.list.FilterState() == widgetlist.Filtering {
		return cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.UKey):
			if bubblehelp.IsKeybindVisible(keybind.UKey) {
				if s.checkRunningTask() {
					switch s.mode {
					case modeInventory:
						s.useItem(index, &selectedItem)
						return cmd
					case modeEquipped:
						s.unequipItem(index, &selectedItem)
						return cmd
					}
				}
			}

		case key.Matches(msg, keybind.EKey):
			if bubblehelp.IsKeybindVisible(keybind.EKey) {
				if s.checkRunningTask() {
					s.equipItem(index, &selectedItem)
					return cmd
				}
			}

		case key.Matches(msg, keybind.Tab):
			s.changeMode()
			return cmd
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

	s.stackCountTitle.SetValue(fmt.Sprintf("%s (%d / %d)",
		lokyn.L("Stacks"), inventory.StacksCount, inventory.MaxStacks))
}

func (s *Screen) loadEquippedInventory() {
	var stacks []api.StackResponse

	resp, err := helper.SendRequest(request.InventoryGetEquippedItems())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	equippedItems := *resp.Result().(*[]api.ItemResponse)

	for _, item := range equippedItems {
		stack := api.StackResponse{
			ID:     0,
			ItemID: item.ID,
			Item:   item,
			Count:  0,
		}

		stacks = append(stacks, stack)
	}

	s.list.SetItems(stacks)
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
		s.removeItem(index)
		return
	}

	s.list.SetItem(index, *item)
}

func (s *Screen) equipItem(index int, item *api.StackResponse) {
	req := request.InventoryEquipItem(item.ItemID)

	_, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	item.Count--

	s.statusMessage.SetMessage(lokyn.L("Item equipped !"), statusmessage.SuccessMessage)

	if item.Count == 0 {
		s.removeItem(index)
		return
	}

	s.list.SetItem(index, *item)
}

func (s *Screen) unequipItem(index int, item *api.StackResponse) {
	req := request.InventoryUnequipItem(item.ItemID)

	_, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.removeItem(index)

	s.statusMessage.SetMessage(lokyn.L("Item unequipped !"), statusmessage.SuccessMessage)
}

func (s *Screen) removeItem(index int) {
	s.list.RemoveItem(index)

	selectedItem := s.list.GetSelectedItem()

	s.updateKeybind(&selectedItem.Item)
}

func (s *Screen) updateKeybind(item *api.ItemResponse) {
	if item == nil || item.ID == 0 {
		bubblehelp.SetKeybindVisible(keybind.UKey, false)
		bubblehelp.SetKeybindVisible(keybind.EKey, false)
		return
	}

	if item.IsUsable {
		bubblehelp.SetKeybindVisible(keybind.UKey, true)
	} else {
		bubblehelp.SetKeybindVisible(keybind.UKey, false)
	}

	if item.EquipmentSlot != nil && s.mode == modeInventory {
		bubblehelp.SetKeybindVisible(keybind.EKey, true)
	} else {
		bubblehelp.SetKeybindVisible(keybind.EKey, false)
	}

	switch s.mode {
	case modeInventory:
		bubblehelp.UpdateKeybindHelpDesc(keybind.UKey, "") // Default
		bubblehelp.UpdateKeybindHelpDesc(keybind.Tab, "")  // Default
	case modeEquipped:
		bubblehelp.SetKeybindVisible(keybind.UKey, true)
		bubblehelp.UpdateKeybindHelpDesc(keybind.UKey, lokyn.L("unequip item"))
		bubblehelp.UpdateKeybindHelpDesc(keybind.Tab, lokyn.L("inventory"))
	}
}

func (s *Screen) changeMode() {
	s.list.BlurCurrent()

	switch s.mode {
	case modeInventory: // goto equipped mode
		s.title.SetValue(s.equippedTitle)
		s.loadEquippedInventory()
		s.mode = modeEquipped
	case modeEquipped: // goto inventory mode
		s.title.SetValue(s.inventoryTitle)
		s.loadInventory()
		s.mode = modeInventory
	}

	s.list.FocusFirst()

	selectedItem := s.list.GetSelectedItem()

	s.updateInspector(&selectedItem)
	s.updateKeybind(&selectedItem.Item)
}

func (s *Screen) checkRunningTask() bool {
	if context.RunningTask != nil {
		s.statusMessage.SetMessage(lokyn.L("A task is running. Claim the reward before doing this."), statusmessage.InformationMessage)
		return false
	}

	return true
}
