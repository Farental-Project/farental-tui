package mailattachmentselect

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/internal/orvyn"
	"farental/layout"
	"farental/style"
	"farental/widget/filterablelist"
	"github.com/charmbracelet/bubbles/key"
	tealist "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
)

type HideAttachmentSelectMsg uint

func HideAttachmentSelectCmd() tea.Msg {
	return HideAttachmentSelectMsg(1)
}

type SelectItemMsg struct {
	Item   api.ItemResponse
	Amount int
}

func SelectItemCmd(item *api.ItemResponse, amount int) tea.Cmd {
	return func() tea.Msg {
		return SelectItemMsg{
			Item:   *item,
			Amount: amount,
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

	w.title = orvyn.NewSimpleRenderable(lokyn.L("Inventory"))
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

				return SelectItemCmd(&selectedItem.Item,
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

func (w *Widget) SetItems(items *[]ListItem) {
	listItems := make([]tealist.Item, 0)

	for _, i := range *items {
		listItems = append(listItems, i)
	}

	w.list.SetItems(listItems)
}

func (w *Widget) LoadData(filterItems []ListItem) {
	var items []ListItem

	w.Init()

	items = make([]ListItem, 0)

	resp, err := helper.SendRequest(request.InventoryGetShareable())

	if err != nil {
		return
	}

	inventory := *resp.Result().(*api.InventoryResponse)

	for _, s := range inventory.Stacks {
		index := FindItemIndex(s.ItemID, &items)

		// Non-existing index
		if index == -1 {
			listItem := NewListItem(&s)
			items = append(items, listItem)
			continue
		}

		items[index].Count += s.Count
	}

	w.filterItems(&items, filterItems)

	w.SetItems(&items)
}

func (w *Widget) filterItems(items *[]ListItem, filterItems []ListItem) {
	tmpItems := *items

	for i, f := range *items {
		index := FindItemIndex(f.Item.ID, &filterItems)

		if index == -1 {
			continue
		}

		tmpItems[i].Count -= filterItems[index].Count
	}

	for i := len(tmpItems) - 1; i >= 0; i-- {
		if tmpItems[i].Count <= 0 {
			tmpItems = helper.SliceRemove(tmpItems, i)
		}
	}

	items = &tmpItems

}

func FindItemIndex(itemID uint, items *[]ListItem) int {
	for i, item := range *items {
		if item.Item.ID == itemID {
			return i
		}
	}

	return -1
}
