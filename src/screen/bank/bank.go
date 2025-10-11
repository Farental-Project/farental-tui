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
	"farental/widget/inventorylistitem"
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

	title              *orvyn.SimpleRenderable
	noBankAccountTitle *orvyn.SimpleRenderable

	characterInventoryList *list.Widget[api.StackResponse]
	bankInventoryList      *list.Widget[api.StackResponse]

	statusMessage *statusmessage.Widget

	help *help.Widget

	focusManager *orvyn.FocusManager

	layout          *layout.CenterLayout
	layoutNoAccount *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable(lokyn.L("Bank"))
	s.title.Style = t.Style(theme.TitleStyleID)

	s.noBankAccountTitle = orvyn.NewSimpleRenderable("NO BANK ACCOUNT")
	s.noBankAccountTitle.Style = t.Style(theme.TitleStyleID)

	s.characterInventoryList = list.New(inventorylistitem.Constructor)
	s.characterInventoryList.PreferredSize.Width = t.Size(ftheme.LayoutWidthSizeID)
	s.characterInventoryList.PreferredSize.Height = 80
	s.characterInventoryList.MinSize.Height = 13

	s.bankInventoryList = list.New(inventorylistitem.Constructor)
	s.bankInventoryList.PreferredSize.Width = t.Size(ftheme.LayoutWidthSizeID)
	s.bankInventoryList.PreferredSize.Height = 80
	s.bankInventoryList.MinSize.Height = 13

	s.statusMessage = statusmessage.New()

	s.help = help.New()

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.characterInventoryList)
	s.focusManager.Add(s.bankInventoryList)

	listsLayout := layout.NewHBoxFixedRatioLayout(0, 1,
		0,
		[]layout.FixedRatioRenderable{
			layout.NewFixedRatioRenderable(0.50, s.characterInventoryList),
			layout.NewFixedRatioRenderable(0.50, s.bankInventoryList),
		},
	)

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(10, 4), 2,
			[]orvyn.Renderable{
				s.title,
				orvyn.VGap,
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
	var inventory api.InventoryResponse

	resp, err := helper.SendRequest(request.InventoryGetFull())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	inventory = *resp.Result().(*api.InventoryResponse)

	s.characterInventoryList.SetItems(inventory.Stacks)
}

func (s *Screen) loadBankAccount() {
	var inventory api.InventoryResponse

	resp, err := helper.SendRequest(request.LocationBankGetFull())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	inventory = *resp.Result().(*api.InventoryResponse)

	s.bankInventoryList.SetItems(inventory.Stacks)
}

func (s *Screen) transfertItem() {
	var item api.StackResponse
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
		req = request.LocationBankTransferTo(item.ItemID, 1)
	} else {
		req = request.LocationBankTransferFrom(item.ItemID, 1)
	}

	resp, err := helper.SendRequest(req)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	if resp.StatusCode() == 200 {
		s.statusMessage.SetMessage(lokyn.L("Item successfully transferred !"), statusmessage.SuccessMessage)
		s.characterInventoryList.Init()
		s.bankInventoryList.Init()
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
