package mailattachmentselect

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/widget/mailattachmentselectlistitem"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
)

type HideAttachmentSelectMsg uint

func HideAttachmentSelectCmd() tea.Msg {
	return HideAttachmentSelectMsg(1)
}

type SelectItemMsg struct {
	Item   api.ItemResponse
	Amount int
}

func SelectItemCmd(data *mailattachmentselectlistitem.Data) tea.Cmd {
	return func() tea.Msg {
		return SelectItemMsg{
			Item:   data.ItemResponse,
			Amount: data.Amount,
		}
	}
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	title *orvyn.SimpleRenderable

	list *list.Widget[mailattachmentselectlistitem.Data]

	layout *layout.VBoxFullLayout

	contentSize orvyn.Size
}

func New() *Widget {
	w := new(Widget)

	t := orvyn.GetTheme()

	w.BaseWidget = orvyn.NewBaseWidget()

	w.title = orvyn.NewSimpleRenderable(lokyn.L("Inventory"))
	w.title.Style = t.Style(ftheme.DimUnderlinedTextStyleID)
	w.title.SizeConstraint = true

	w.list = list.New(mailattachmentselectlistitem.Constructor)

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
			if w.list.FilterState() != list.Filtering {
				selectedItem := w.list.GetSelectedItem()

				if selectedItem.Amount > 0 {
					return SelectItemCmd(&selectedItem)
				}
			}

		case key.Matches(msg, keybind.Esc):
			if w.list.FilterState() == list.Unfiltered {
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
	return orvyn.GetTheme().Style(theme.BlurredWidgetStyleID).
		Width(w.contentSize.Width).
		Height(w.contentSize.Height).
		Render(w.layout.Render())
}

func (w *Widget) Resize(size orvyn.Size) {
	s := orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)

	w.BaseWidget.Resize(size)

	size.Width -= s.GetHorizontalFrameSize()
	size.Height -= s.GetVerticalFrameSize()

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
	return w.list.FilterState() == list.Unfiltered
}

func (w *Widget) LoadData(filterItems []mailattachmentselectlistitem.Data) {
	w.Init()

	items := make([]mailattachmentselectlistitem.Data, 0)

	resp, err := helper.SendRequest(request.InventoryGetShareable())

	if err != nil {
		return
	}

	inventory := *resp.Result().(*api.InventoryResponse)

	for _, s := range inventory.Stacks {
		index := FindItemIndex(s.ItemID, &items)

		// Non-existing index
		if index == -1 {
			listItem := mailattachmentselectlistitem.Data{
				ItemResponse: s.Item,
				Count:        s.Count,
				Amount:       0,
			}

			items = append(items, listItem)
			continue
		}

		items[index].Count += s.Count
	}

	w.filterItems(&items, filterItems)

	w.list.SetItems(items)
}

func (w *Widget) filterItems(items *[]mailattachmentselectlistitem.Data, filterItems []mailattachmentselectlistitem.Data) {
	tmpItems := *items

	for i, f := range *items {
		index := FindItemIndex(f.ID, &filterItems)

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

func FindItemIndex(itemID uint, items *[]mailattachmentselectlistitem.Data) int {
	for i, item := range *items {
		if item.ID == itemID {
			return i
		}
	}

	return -1
}
