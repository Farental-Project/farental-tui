package mailattachmentselect

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/lang"
	"farental/internal/orvyn"
	"farental/layout"
	"farental/style"
	"farental/widget/filterablelist"
	"github.com/charmbracelet/bubbles/key"
	tealist "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
)

type HideAttachmentSelectMsg uint

func HideAttachmentSelectCmd() tea.Msg {
	return HideAttachmentSelectMsg(1)
}

type SelectItemMsg struct {
	StackID  uint
	ItemID   uint
	ItemName string
	Amount   int
}

func SelectItemCmd(stackID uint, itemID uint, name string, amount int) tea.Cmd {
	return func() tea.Msg {
		return SelectItemMsg{
			StackID:  stackID,
			ItemID:   itemID,
			ItemName: name,
			Amount:   amount,
		}
	}
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	title *orvyn.SimpleRenderable

	list *filterablelist.Widget

	layout *layout.VBoxFullLayout

	contentSize orvyn.Size
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.title = orvyn.NewSimpleRenderable(lang.L("Inventory"))
	w.title.Style = style.DimUnderlinedTitleStyle
	w.title.SizeConstraint = true

	w.list = filterablelist.New(ListItemDelegate{}, []tealist.Item{})
	w.list.KeyMap.NextPage = keybind.NextPage
	w.list.KeyMap.PrevPage = keybind.PrevPage

	w.layout = layout.NewMaxWidthVBoxFullLayout(
		orvyn.NewSize(0, 0),
		1,
		[]orvyn.Renderable{
			w.title,
			w.list,
		},
	)

	return w
}

func (w *Widget) Init() tea.Cmd {
	return w.list.Init()
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Enter):
			if w.list.FilterState() != tealist.Filtering {
				selectedItem, ok := w.list.SelectedItem().(ListItem)

				if !ok {
					return nil
				}

				return SelectItemCmd(selectedItem.Stack.ID,
					selectedItem.Stack.ItemID,
					selectedItem.Stack.Item.Name,
					selectedItem.Amount)
			}

		case key.Matches(msg, keybind.Esc):
			if w.list.FilterState() == tealist.Unfiltered {
				return HideAttachmentSelectCmd
			}

		case key.Matches(msg, keybind.Help):
			bubblehelp.ShowAll = !bubblehelp.ShowAll

			return nil
		}
	}

	cmd := w.list.Update(msg)

	return cmd
}

func (w *Widget) Render() string {
	var s lipgloss.Style

	if w.IsFocused() {
		s = style.FocusedStyle
	} else {
		s = style.BlurredStyle
	}

	return s.Width(w.contentSize.Width).
		Height(w.contentSize.Height).
		Render(w.layout.Render())
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	size.Width -= style.BlurredStyle.GetHorizontalFrameSize()
	size.Height -= style.BlurredStyle.GetVerticalFrameSize()

	w.contentSize = size
	w.layout.Resize(size)
}

func (w *Widget) OnFocus() {
}

func (w *Widget) OnBlur() {
}

func (w *Widget) OnEnterInput() {
	bubblehelp.SwitchContext(keybind.ContextFilterSelectionListIncDec)
}

func (w *Widget) OnExitInput() {
	bubblehelp.SwitchToPreviousContext()
}

func (w *Widget) CanExitInputting() bool {
	return w.list.FilterState() == tealist.Unfiltered
}

func (w *Widget) LoadData(filterItems []ListItem) {
	var items []tealist.Item

	w.Init()

	items = make([]tealist.Item, 0)

	resp, err := helper.SendRequest(request.InventoryGetShareable())

	if err != nil {
		return
	}

	inventory := *resp.Result().(*api.InventoryResponse)

	for _, s := range inventory.Stacks {
		item := ListItem{
			Stack: s,
		}

		filter := w.filterStack(&item, &filterItems)

		if !filter {
			items = append(items, item)
		}
	}

	w.list.SetItems(items)
}

func (w *Widget) filterStack(item *ListItem, filterItems *[]ListItem) bool {
	var filterItem *ListItem

	for _, f := range *filterItems {
		if f.Stack.ID == item.Stack.ID {
			filterItem = &f
			break
		}
	}

	if filterItem == nil {
		return false
	}

	item.Stack.Count -= filterItem.Amount

	if item.Stack.Count <= 0 {
		return true
	}

	return false
}
