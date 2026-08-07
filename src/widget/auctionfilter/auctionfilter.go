package auctionfilter

import (
	"farental/core/data/api"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/screen/dialog/statskillselection"
	"farental/widget/button"
	"farental/widget/multivalueselector"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/widget/label"
	"github.com/halsten-dev/orvyn/widget/textinput"
)

// dialogIDStatSkill is opened by the stat/skill button and read back in Update.
const dialogIDStatSkill orvyn.ScreenID = "auctionStatSkill"

// searchIndex is the search box's position in the panel's focus manager. The
// screen asks whether the cursor sits there before it claims a letter key.
const searchIndex = 0

// AppliedMsg tells the screen the player asked for the current filter to be
// run. The widget never queries anything itself.
type AppliedMsg int

func AppliedCmd() tea.Msg {
	return AppliedMsg(1)
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	lblSearch *label.Widget
	tiSearch  *textinput.Widget

	lblScope *label.Widget
	mvsScope *multivalueselector.Widget[Option]

	lblKind      *label.Widget
	lblSlot      *label.Widget
	lblWeapon    *label.Widget
	lblStatSkill *label.Widget
	lblMinStat   *label.Widget

	mvsKind   *multivalueselector.Widget[Option]
	mvsSlot   *multivalueselector.Widget[Option]
	mvsWeapon *multivalueselector.Widget[Option]

	btStatSkill *button.Widget
	tiMinStat   *textinput.Widget

	focusManager *orvyn.FocusManager

	options *api.AuctionFilterOptionsResponse

	// statCode and skillCode are mutually exclusive by construction: the picker
	// returns one entry, and setting one clears the other.
	statCode  string
	skillCode string

	layout *layout.VBoxLayout
}

var _ orvyn.Widget = (*Widget)(nil)
var _ orvyn.Focusable = (*Widget)(nil)

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.lblSearch = label.New("")

	// No Validate: this one takes any text the player types.
	w.tiSearch = textinput.New()

	w.lblScope = label.New("")

	w.mvsScope = multivalueselector.New[Option]()
	w.mvsScope.OnBlur()

	w.lblKind = label.New("")
	w.lblSlot = label.New("")
	w.lblWeapon = label.New("")
	w.lblStatSkill = label.New("")
	w.lblMinStat = label.New("")

	w.mvsKind = multivalueselector.New[Option]()
	w.mvsKind.OnBlur()

	w.mvsSlot = multivalueselector.New[Option]()
	w.mvsSlot.OnBlur()

	w.mvsWeapon = multivalueselector.New[Option]()
	w.mvsWeapon.OnBlur()

	w.btStatSkill = button.New("")
	w.btStatSkill.OnClickedCallback = w.btStatSkillOnClicked
	w.btStatSkill.OnFocusCallback = w.btOnFocus
	w.btStatSkill.OnBlurCallback = w.btOnBlur

	w.tiMinStat = textinput.New()
	w.tiMinStat.Validate = helper.SignedNumericalValidate

	w.focusManager = orvyn.NewFocusManager()
	w.focusManager.Add(w.tiSearch)
	w.focusManager.Add(w.mvsScope)
	w.focusManager.Add(w.mvsKind)
	w.focusManager.Add(w.mvsSlot)
	w.focusManager.Add(w.mvsWeapon)
	w.focusManager.Add(w.btStatSkill)
	w.focusManager.Add(w.tiMinStat)

	// ↑/↓ inside the panel leaves Tab to the screen, which uses it to move
	// between the panel and the auction list.
	w.focusManager.NextFocusKeybind = keybind.Down
	w.focusManager.PreviousFocusKeybind = keybind.Up

	w.layout = layout.NewMaxWidthVBoxLayout(0,
		w.lblSearch,
		w.tiSearch,
		w.lblScope,
		w.mvsScope,
		w.lblKind,
		w.mvsKind,
		w.lblSlot,
		w.mvsSlot,
		w.lblWeapon,
		w.mvsWeapon,
		w.lblStatSkill,
		w.btStatSkill,
		w.lblMinStat,
		w.tiMinStat,
	)

	w.Reset()

	return w
}

func (w *Widget) Init() tea.Cmd {
	// Filled before Reset: it selects index 0, and an empty selector has none.
	// These two are client-side strings rather than server vocabulary, but
	// they are localized just the same, so they are refreshed on every entry.
	setOptions(w.mvsScope, []Option{
		{Code: "", Label: lokyn.L("All auctions")},
		{Code: scopeOutbid, Label: lokyn.L("My outbid auctions")},
	})

	// Reset here, not in SetOptions: a failed options fetch skips SetOptions
	// entirely, and the screen still needs an empty filter to apply.
	w.Reset()

	w.lblSearch.SetValue(lokyn.L("Search"))
	w.lblScope.SetValue(lokyn.L("Show"))
	w.lblKind.SetValue(lokyn.L("Kind"))
	w.lblSlot.SetValue(lokyn.L("Equipment slot"))
	w.lblWeapon.SetValue(lokyn.L("Weapon type"))
	w.lblStatSkill.SetValue(lokyn.L("Stat or skill"))
	w.lblMinStat.SetValue(lokyn.L("Minimum"))

	w.tiSearch.Placeholder = lokyn.L("Filter loaded auctions")
	w.tiMinStat.Placeholder = lokyn.L("Minimum value")

	w.focusManager.FocusFirst()

	return nil
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case orvyn.DialogExitMsg:
		if msg.DialogID == dialogIDStatSkill {
			// The dialog switches context on entry and never switches back on its
			// own; restore it on every exit, cancel included.
			bubblehelp.SwitchToPreviousContext()

			// Restoring resets the dialog's own keymap, not ours, but Reset() on
			// entry set every binding in ours back to visible too — re-assert
			// Space's hidden-when-blurred state in case the button still has focus.
			if w.btStatSkill.IsFocused() {
				bubblehelp.SetKeybindVisible(keybind.Space, true)
			}

			entry, ok := msg.Param.(statskillselection.Entry)

			if ok {
				w.statSkillSelected(entry)
			}

			return nil
		}
	}

	if k, ok := orvyn.GetKeyMsg(msg); ok {
		if key.Matches(k, keybind.Enter) && w.IsFocused() {
			return AppliedCmd
		}
	}

	return w.focusManager.Update(msg)
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	w.layout.Resize(w.GetContentSize())
}

func (w *Widget) Render() string {
	contentSize := w.GetContentSize()

	return w.GetStyle().Width(contentSize.Width).
		Height(contentSize.Height).
		Render(w.layout.Render())
}

func (w *Widget) OnFocus() {
	w.BaseFocusable.OnFocus()
	w.focusManager.FocusFirst()
}

func (w *Widget) OnBlur() {
	w.BaseFocusable.OnBlur()
	w.focusManager.BlurCurrent()
}

// SetOptions fills the selectors with the server's vocabulary. Called on every
// screen entry, since the labels follow the player's language.
func (w *Widget) SetOptions(options *api.AuctionFilterOptionsResponse) {
	w.options = options

	kinds := []Option{{Label: lokyn.L("Any kind")}}
	slots := []Option{{Label: lokyn.L("Any slot")}}
	weapons := []Option{{Label: lokyn.L("Any weapon")}}

	if options != nil {
		for _, kind := range options.Kinds {
			kinds = append(kinds, Option{Code: kind.Code, Label: kind.Name})
		}

		for _, slot := range options.EquipmentSlots {
			slots = append(slots, Option{Code: slot.Code, Label: slot.Name})
		}

		for _, weaponType := range options.WeaponTypes {
			weapons = append(weapons, Option{Code: weaponType.Code, Label: weaponType.Name})
		}
	}

	setOptions(w.mvsKind, kinds)
	setOptions(w.mvsSlot, slots)
	setOptions(w.mvsWeapon, weapons)
}

// setOptions feeds a selector, which keys its values by the label it renders.
func setOptions(selector *multivalueselector.Widget[Option], options []Option) {
	keys := make([]string, 0, len(options))
	values := make(map[string]Option, len(options))

	for _, option := range options {
		keys = append(keys, option.Label)
		values[option.Label] = option
	}

	selector.SetValues(keys, values)
}

// Reset returns every control to "no filter". It does not move the cursor:
// pressing r from the auction list calls this too, and jumping focus into the
// panel from there would style the Scope selector as focused while the list
// still has real focus. Callers that need the cursor moved do it themselves
// (see Init).
func (w *Widget) Reset() {
	w.mvsScope.SetSelected(0)
	w.mvsKind.SetSelected(0)
	w.mvsSlot.SetSelected(0)
	w.mvsWeapon.SetSelected(0)

	w.statCode = ""
	w.skillCode = ""

	w.tiSearch.SetValue("")
	w.btStatSkill.SetLabel(lokyn.L("Any stat or skill"))
	w.tiMinStat.SetValue("")
}

// GetFilter reads the controls into the query the request layer sends.
func (w *Widget) GetFilter() api.AuctionFilter {
	return buildFilter(
		w.mvsScope.GetSelectedValue().Code,
		w.mvsKind.GetSelectedValue().Code,
		w.mvsSlot.GetSelectedValue().Code,
		w.mvsWeapon.GetSelectedValue().Code,
		w.statCode,
		w.skillCode,
		w.tiMinStat.Value(),
	)
}

// GetSearch reads the local search box. It stays out of GetFilter: the text
// narrows rows already on screen instead of reaching the server.
func (w *Widget) GetSearch() string {
	return w.tiSearch.Value()
}

// SearchFocused is true only when the panel is focused and its cursor sits
// on the search box; the screen uses it to stand down letter keybinds.
func (w *Widget) SearchFocused() bool {
	return w.IsFocused() && w.focusManager.TabIndex() == searchIndex
}

func (w *Widget) btOnFocus() {
	bubblehelp.SetKeybindVisible(keybind.Space, true)
}

func (w *Widget) btOnBlur() {
	bubblehelp.SetKeybindVisible(keybind.Space, false)
}

func (w *Widget) btStatSkillOnClicked() tea.Cmd {
	return orvyn.OpenDialog(dialogIDStatSkill, statskillselection.New(w.options), nil)
}

func (w *Widget) statSkillSelected(entry statskillselection.Entry) {
	w.statCode = ""
	w.skillCode = ""

	switch entry.Kind {
	case statskillselection.EntryStat:
		w.statCode = entry.Code
	case statskillselection.EntrySkill:
		w.skillCode = entry.Code
	}

	if entry.Kind == statskillselection.EntryNone {
		w.btStatSkill.SetLabel(lokyn.L("Any stat or skill"))
		return
	}

	w.btStatSkill.SetLabel(entry.Label)
}
