package shop

import (
	"farental/art"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen"
	"farental/screen/dialog/popup"
	"farental/widget/help"
	"farental/widget/inventoryshoplistitem"
	"fmt"
	"net/http"

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
	title *orvyn.SimpleRenderable

	inventoryList *list.Widget[inventoryshoplistitem.Data]

	statusMessage *statusmessage.Widget

	help *help.Widget

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable(lokyn.L("Shop"))
	s.title.Style = t.Style(theme.TitleStyleID)

	s.inventoryList = list.New(inventoryshoplistitem.Constructor)

	s.statusMessage = statusmessage.New()

	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(10, 4), 2,
			s.title,
			orvyn.VGap,
			s.inventoryList,
			s.statusMessage,
			s.help,
		),
	)

	return s
}

func (s *Screen) OnEnter(_ any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextShop)

	s.inventoryList.Init()

	s.inventoryList.OnFocus()

	s.loadInventory()

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(msg, keybind.Enter):
			item := s.inventoryList.GetSelectedItem()

			sellPrice := item.Amount * item.SellPrice

			orvyn.OpenDialog("sellItems", popup.NewYesNo(
				fmt.Sprintf(lokyn.L("Are sure you want to sell %dx %s for a total of %d%c ?"),
					item.Amount, item.Name, sellPrice, art.CharGrynars),
			), nil)

		case key.Matches(msg, keybind.Esc):
			return orvyn.SwitchScreen(screen.IDDashBoard)

		}
	}

	switch msg := msg.(type) {
	case orvyn.DialogExitMsg:
		switch msg.DialogID {
		case "sellItems":
			val := msg.Param.(uint)

			switch val {
			case 1:
				s.sellItems()
			}
		}
	}

	cmd := s.inventoryList.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) loadInventory() {
	resp, err := helper.SendRequest(request.InventoryGetSellable())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	inventory := resp.Result().(*api.InventoryResponse)

	s.inventoryList.SetItems(s.initListItems(inventory))
}

func (s *Screen) sellItems() {
	item := s.inventoryList.GetSelectedItem()

	resp, err := helper.SendRequest(request.LocationMerchantSellItem(item.ID, item.Amount))

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	if resp.StatusCode() == http.StatusOK {
		s.statusMessage.SetMessage(lokyn.L("Items successfully sold !"), statusmessage.SuccessMessage)
		s.loadInventory()
	}
}

func (s *Screen) initListItems(inventory *api.InventoryResponse) []inventoryshoplistitem.Data {
	var listItemsData []inventoryshoplistitem.Data

	listItemsData = make([]inventoryshoplistitem.Data, 0)

	for _, s := range inventory.Stacks {
		index := findItemIndex(s.ItemID, &listItemsData)

		if index == -1 {
			listItem := inventoryshoplistitem.Data{
				ItemResponse: s.Item,
				Count:        s.Count,
				Amount:       0,
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
