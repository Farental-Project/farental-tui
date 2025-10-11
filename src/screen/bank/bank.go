package bank

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
	"farental/widget/inventorygroupedlistitem"
	"fmt"
	"net/http"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-resty/resty/v2"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
)

type Screen struct {
	existingAccount bool

	// title              *orvyn.SimpleRenderable
	maxStackInfo       *orvyn.SimpleRenderable
	noBankAccountTitle *orvyn.SimpleRenderable

	characterInventoryTitle *orvyn.SimpleRenderable
	characterInventoryList  *list.Widget[inventorygroupedlistitem.Data]

	bankInventoryTitle *orvyn.SimpleRenderable
	bankInventoryList  *list.Widget[inventorygroupedlistitem.Data]

	statusMessage *statusmessage.Widget

	help *help.Widget

	focusManager *orvyn.FocusManager

	layout          *layout.CenterLayout
	layoutNoAccount *layout.CenterLayout

	maxStackCount int
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()
	ts := t.Style(theme.TitleStyleID)

	// s.title = orvyn.NewSimpleRenderable(lokyn.L("Bank"))
	// s.title.Style = t.Style(theme.TitleStyleID)

	s.maxStackInfo = orvyn.NewSimpleRenderable("")
	s.maxStackInfo.Style = t.Style(theme.DimTextStyleID)

	s.noBankAccountTitle = orvyn.NewSimpleRenderable("NO BANK ACCOUNT")
	s.noBankAccountTitle.Style = ts

	s.characterInventoryTitle = orvyn.NewSimpleRenderable(lokyn.L("Inventaire"))
	s.characterInventoryTitle.SizeConstraint = true
	s.characterInventoryTitle.Style = ts

	s.characterInventoryList = list.New(inventorygroupedlistitem.Constructor)
	s.characterInventoryList.PreferredSize.Width = t.Size(ftheme.LayoutWidthSizeID)
	s.characterInventoryList.PreferredSize.Height = 80
	s.characterInventoryList.MinSize.Height = 13

	s.bankInventoryTitle = orvyn.NewSimpleRenderable("")
	s.bankInventoryTitle.SizeConstraint = true
	s.bankInventoryTitle.Style = ts

	s.bankInventoryList = list.New(inventorygroupedlistitem.Constructor)
	s.bankInventoryList.PreferredSize.Width = t.Size(ftheme.LayoutWidthSizeID)
	s.bankInventoryList.PreferredSize.Height = 80
	s.bankInventoryList.MinSize.Height = 13

	s.statusMessage = statusmessage.New()

	s.help = help.New()

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.characterInventoryList)
	s.focusManager.Add(s.bankInventoryList)

	characterListLayout := layout.NewMaxWidthVBoxFullLayout(
		orvyn.NewSize(0, 0), 1,
		[]orvyn.Renderable{
			s.characterInventoryTitle,
			s.characterInventoryList,
		})

	bankListLayout := layout.NewMaxWidthVBoxFullLayout(
		orvyn.NewSize(0, 0), 1,
		[]orvyn.Renderable{
			s.bankInventoryTitle,
			s.bankInventoryList,
		})

	listsLayout := layout.NewHBoxFixedRatioLayout(0, 1,
		1,
		[]layout.FixedRatioRenderable{
			layout.NewFixedRatioRenderable(0.50, characterListLayout),
			layout.NewFixedRatioRenderable(0.50, bankListLayout),
		},
	)

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(10, 4), 0,
			[]orvyn.Renderable{
				// s.title,
				// s.maxStackInfo,
				// orvyn.VGap,
				listsLayout,
				s.statusMessage,
				s.help,
			},
		),
	)

	s.layoutNoAccount = layout.NewCenterLayout(
		s.noBankAccountTitle,
	)

	return s
}

func (s *Screen) OnEnter(any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextBank)

	s.characterInventoryList.Init()
	s.bankInventoryList.Init()

	resp, _ := helper.SendRequest(request.CharacterHaveBankAccount())

	if resp.StatusCode() == http.StatusNotFound {
		orvyn.OpenDialog("buyAccount", popup.NewYesNo(
			fmt.Sprintf(lokyn.L("Do you want to open your bank account for 5000%c ?"), art.CharGrynars),
		), nil)

		s.existingAccount = false
	} else {
		s.existingAccount = true
		s.focusManager.FocusFirst()
		s.loadInventory()
		s.loadBankAccount()
	}

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Esc):
			if s.currentListFilterState() == list.Unfiltered {
				return orvyn.SwitchScreen(screen.IDDashBoard)
			}
		case key.Matches(msg, keybind.TKey):
			s.transfertItem()

			return nil
		}

	case orvyn.DialogExitMsg:
		switch msg.DialogID {
		case "buyAccount":
			val := msg.Param.(uint)
			switch val {
			case 1:
				s.openBankAccount()
			default:
				return orvyn.SwitchScreen(screen.IDDashBoard)
			}
		}
	}

	cmd := s.focusManager.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	if !s.existingAccount {
		return s.layoutNoAccount
	}

	return s.layout
}

func (s *Screen) openBankAccount() {
	resp, err := helper.SendRequest(request.LocationCreateBankAccount())

	if err != nil {
		s.statusMessage.SetError(err)
	}

	if resp.StatusCode() == http.StatusCreated {
		s.existingAccount = true
	}
}

func (s *Screen) loadInventory() {
	var inventory *api.InventoryResponse

	resp, err := helper.SendRequest(request.InventoryGetFull())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	inventory = resp.Result().(*api.InventoryResponse)

	listItems := s.initListItems(inventory)

	s.characterInventoryList.SetItems(listItems)
}

func (s *Screen) loadBankAccount() {
	resp, err := helper.SendRequest(request.LocationBankGetAccount())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	bankAccount := resp.Result().(*api.BankAccountResponse)

	listItems := s.initListItems(&bankAccount.Inventory)

	s.bankInventoryList.SetItems(listItems)
	s.maxStackCount = bankAccount.MaxStackCount
	s.bankInventoryTitle.SetValue(fmt.Sprintf(lokyn.L("Bank (Max stack count : %d)"), s.maxStackCount))
}

func (s *Screen) initListItems(inventory *api.InventoryResponse) []inventorygroupedlistitem.Data {
	var listItemsData []inventorygroupedlistitem.Data

	listItemsData = make([]inventorygroupedlistitem.Data, 0)

	for _, s := range inventory.Stacks {
		index := findItemIndex(s.ItemID, &listItemsData)

		if index == -1 {
			listItem := inventorygroupedlistitem.Data{
				ItemResponse: s.Item,
				Count:        s.Count,
				Amount:       0,
				StackCount:   1,
			}

			listItemsData = append(listItemsData, listItem)
			continue
		}

		listItemsData[index].Count += s.Count
		listItemsData[index].StackCount++
	}

	return listItemsData
}

func (s *Screen) transfertItem() {
	var item inventorygroupedlistitem.Data
	var toBank bool
	var req *resty.Request

	tabIndex := s.focusManager.TabIndex()

	switch tabIndex {
	case 0:
		item = s.characterInventoryList.GetSelectedItem()
		toBank = true
	case 1:
		item = s.bankInventoryList.GetSelectedItem()
		toBank = false
	}

	if item.ID == 0 {
		return
	}

	if toBank {
		req = request.LocationBankTransferTo(item.ID, item.Amount)
	} else {
		req = request.LocationBankTransferFrom(item.ID, item.Amount)
	}

	resp, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	if resp.StatusCode() == 200 {
		message := ""

		if toBank {
			message = lokyn.L("Item successfully transferred to the bank !")
		} else {
			message = lokyn.L("Item successfully transferred to the inventory !")
		}

		s.statusMessage.SetMessage(message, statusmessage.SuccessMessage)
		s.loadInventory()
		s.loadBankAccount()
	}
}

func (s *Screen) currentListFilterState() list.FilterState {
	tabIndex := s.focusManager.TabIndex()

	switch tabIndex {
	case 0:
		return s.characterInventoryList.FilterState()
	case 1:
		return s.bankInventoryList.FilterState()
	}

	return list.Unfiltered
}

func findItemIndex(itemID uint, data *[]inventorygroupedlistitem.Data) int {
	for i, item := range *data {
		if item.ID == itemID {
			return i
		}
	}

	return -1
}
