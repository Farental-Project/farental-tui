package shop

import (
	"farental/art"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/screen"
	"farental/screen/dialog/popup"
	"farental/widget/help"
	"farental/widget/inventoryshoplistitem"
	"farental/widget/iteminspect"
	"fmt"
	"net/http"

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

type shopMode uint8

const (
	modeBuy shopMode = iota
	modeSell
)

type Screen struct {
	mode shopMode

	buyTitle  string
	sellTitle string

	title *orvyn.SimpleRenderable

	list *widgetlist.Widget[inventoryshoplistitem.Data]

	inspector *iteminspect.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable(s.buyTitle)
	s.title.Style = t.Style(theme.TitleStyleID)

	s.list = widgetlist.New(inventoryshoplistitem.Constructor)

	s.list.SetPreferredSize(orvyn.NewSize(t.Size(ftheme.LayoutWidthSizeID), 80))
	s.list.SetMinSize(orvyn.NewSize(30, 13))

	s.inspector = iteminspect.New()

	s.statusMessage = statusmessage.New()

	s.help = help.New()

	shopElements := []layout.FixedRatioRenderable{
		layout.NewFixedRatioRenderable(0.60, s.list),
		layout.NewFixedRatioRenderable(0.40, s.inspector),
	}

	shopLayout := layout.NewHBoxFixedRatioLayout(0, 1, 0, shopElements...)

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(10, 4), 2,
			s.title,
			orvyn.VGap,
			shopLayout,
			s.statusMessage,
			s.help,
		),
	)

	return s
}

func (s *Screen) OnEnter(_ any) tea.Cmd {
	s.statusMessage.Reset()

	bubblehelp.SwitchContext(keybind.ContextShop)

	s.buyTitle = lokyn.L("Shop - Buy")
	s.sellTitle = lokyn.L("Shop - Sell")

	s.mode = modeBuy

	s.list.Init()

	s.loadBuyableItems()

	s.list.FocusFirst()

	s.updateInspector()

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := orvyn.GetKeyMsg(msg); ok {
		s.statusMessage.Reset()

		switch {
		case key.Matches(msg, keybind.Tab):
			s.changeMode()
			return nil

		case key.Matches(msg, keybind.Enter):
			confirmMessage := s.getConfirmationMessageFormat()

			item := s.list.GetSelectedItem()

			if item.Amount == 0 {
				return nil
			}

			var price int

			switch s.mode {
			case modeBuy:
				price = item.BuyPrice
			case modeSell:
				price = item.SellPrice
			}

			finalPrice := item.Amount * price

			orvyn.OpenDialog("sellOrBuyItems", popup.NewYesNo(
				fmt.Sprintf(confirmMessage,
					item.Amount, item.Name, finalPrice, art.CharGrynars),
			), nil)

		case key.Matches(msg, keybind.Esc):
			return orvyn.SwitchScreen(screen.IDDashBoard)

		}
	}

	switch msg := msg.(type) {
	case orvyn.DialogExitMsg:
		switch msg.DialogID {
		case "sellOrBuyItems":
			val := msg.Param.(uint)

			switch val {
			case 1:
				switch s.mode {
				case modeBuy:
					s.buyItems()

				case modeSell:
					s.sellItems()
				}
			}
		}
	}

	cmd := s.list.Update(msg)

	s.updateInspector()

	return cmd
}

func (s *Screen) updateInspector() {
	item := s.list.GetSelectedItem()

	if item.ID == 0 {
		return
	}

	if s.inspector.GetCurrentItemID() != item.ID {
		s.inspector.UpdateData(&item.ItemResponse)
	}
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) loadBuyableItems() {
	items, err := helper.Fetch[[]api.ItemResponse](request.LocationMerchantGetBuyableItem())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.list.SetItems(s.initListItems(items))
}

func (s *Screen) loadInventory() {
	inventory, err := helper.Fetch[api.InventoryResponse](request.InventoryGetSellable())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.list.SetItems(s.initListItemsFromInventory(inventory))
}

func (s *Screen) buyItems() {
	item := s.list.GetSelectedItem()

	resp, err := helper.SendRequest(request.LocationMerchantBuyItem(item.ID, item.Amount))

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	if resp.StatusCode() == http.StatusOK {
		s.statusMessage.SetMessage(lokyn.L("Items successfully bought !"),
			statusmessage.SuccessMessage)
		s.loadBuyableItems()
	}
}

func (s *Screen) sellItems() {
	item := s.list.GetSelectedItem()

	resp, err := helper.SendRequest(request.LocationMerchantSellItem(item.ID, item.Amount))

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	if resp.StatusCode() == http.StatusOK {
		s.statusMessage.SetMessage(lokyn.L("Items successfully sold !"),
			statusmessage.SuccessMessage)
		s.loadInventory()
	}
}

func (s *Screen) initListItems(items *[]api.ItemResponse) []inventoryshoplistitem.Data {
	var listItemsData []inventoryshoplistitem.Data

	listItemsData = make([]inventoryshoplistitem.Data, 0)

	for _, i := range *items {
		listItem := inventoryshoplistitem.Data{
			ItemResponse: i,
			Count:        0,
			Amount:       0,
			Buying:       true,
		}

		listItemsData = append(listItemsData, listItem)
	}

	return listItemsData
}

func (s *Screen) initListItemsFromInventory(inventory *api.InventoryResponse) []inventoryshoplistitem.Data {
	var listItemsData []inventoryshoplistitem.Data

	listItemsData = make([]inventoryshoplistitem.Data, 0)

	for _, s := range inventory.Stacks {
		index := findItemIndex(s.ItemID, &listItemsData)

		if index == -1 {
			listItem := inventoryshoplistitem.Data{
				ItemResponse: s.Item,
				Count:        s.Count,
				Amount:       0,
				Buying:       false,
			}

			listItemsData = append(listItemsData, listItem)
			continue
		}

		listItemsData[index].Count += s.Count
	}

	return listItemsData
}

func findItemIndex(itemID uint, data *[]inventoryshoplistitem.Data) int {
	for i, item := range *data {
		if item.ID == itemID {
			return i
		}
	}

	return -1
}

func (s *Screen) getConfirmationMessageFormat() string {
	var format string

	switch s.mode {
	case modeBuy:
		format = lokyn.L("Are you sure you want to buy %dx %s for a total of %d%c ?")

	case modeSell:
		format = lokyn.L("Are you sure you want to sell %dx %s for a total of %d%c ?")
	}

	return format
}

func (s *Screen) updateKeybind(item *inventoryshoplistitem.Data) {
	switch s.mode {
	case modeBuy:
		bubblehelp.UpdateKeybindHelpDesc(keybind.Enter, lokyn.L("Buy"))
		bubblehelp.UpdateKeybindHelpDesc(keybind.Tab, lokyn.L("Sell"))
	case modeSell:
		bubblehelp.UpdateKeybindHelpDesc(keybind.Enter, lokyn.L("Sell"))
		bubblehelp.UpdateKeybindHelpDesc(keybind.Tab, lokyn.L("Buy"))
	}
}

func (s *Screen) changeMode() {
	s.list.BlurCurrent()

	switch s.mode {
	case modeBuy: // goto sell mode
		s.title.SetValue(s.sellTitle)
		s.loadInventory()
		s.mode = modeSell
	case modeSell: // goto buy mode
		s.title.SetValue(s.buyTitle)
		s.loadBuyableItems()
		s.mode = modeBuy
	}

	s.list.FocusFirst()

	selectedItem := s.list.GetSelectedItem()

	s.updateKeybind(&selectedItem)
	s.updateInspector()
}
