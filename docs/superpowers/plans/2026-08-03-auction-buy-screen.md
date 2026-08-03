# Auction house buy screen Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the auction house buy screen — filter the listings with the server's filter vocabulary, load more pages on demand, inspect, bid, or buy outright.

**Architecture:** The screen follows `screen/scripteditor`: sibling focusable widgets in a screen-level `orvyn.FocusManager`, widgets reporting upward with typed `tea.Msg` values. A filter panel (`widget/auctionfilter`) sits left at 0.25 of the width, the auction `widgetlist` right at 0.75. All server calls go through `internal/helper` and the already-built `core/request` auction functions.

**Tech Stack:** Go, bubbletea, orvyn (`github.com/halsten-dev/orvyn`), lokyn (`lokyn.L` translations), bubblehelp (key help contexts), resty via `core/request`.

Spec: `docs/superpowers/specs/2026-08-03-auction-buy-screen-design.md`

## Global Constraints

- Work only in `/home/halsten/Dev/Farental/farental-tui`. No server (`farental-cli`) changes: the API is merged and final.
- Every user-visible string goes through `lokyn.L("…")`.
- Comments explain **why**, not what, and stay short. No comment restating the line below it.
- Tests only where the logic is pure. Widget constructors call `orvyn.GetTheme()` and are not unit-tested.
- Verification for every task, run from `/home/halsten/Dev/Farental/farental-tui/src`:
  `gofmt -l . && go build ./... && go vet ./... && go test ./...`
- Money is rendered with `art.CharGrynars` (a rune, use `%c`).
- List item widgets set `size.Height = 3` in `Resize` — the repo's minimum, one content line inside the border.
- Commit after each task. Conventional commits, matching the repo (`feat:`, `fix:`, `docs:`).

---

### Task 1: Signed numeric input validation

The min-stat field must accept `-5`: stat penalties are legal content, and a filter that cannot express them hides items. `helper.NumericalValidate` is digits-only and is used by other forms, so it gets a sibling rather than a change.

**Files:**
- Modify: `internal/helper/inputvalidation.go`
- Test: `internal/helper/inputvalidation_test.go` (create)

**Interfaces:**
- Consumes: nothing.
- Produces: `helper.SignedNumericalValidate(s string) error`.

- [ ] **Step 1: Write the failing test**

Create `internal/helper/inputvalidation_test.go`:

```go
package helper

import "testing"

func TestSignedNumericalValidate(t *testing.T) {
	valid := []string{"", "0", "5", "-5", "-"}

	for _, s := range valid {
		if err := SignedNumericalValidate(s); err != nil {
			t.Errorf("SignedNumericalValidate(%q) = %v, want nil", s, err)
		}
	}

	invalid := []string{"a", "5-", "--5", "1.5"}

	for _, s := range invalid {
		if err := SignedNumericalValidate(s); err == nil {
			t.Errorf("SignedNumericalValidate(%q) = nil, want an error", s)
		}
	}
}
```

`"-"` is accepted on purpose: it is a minus sign mid-typing. It parses to nothing later, which leaves the minimum unset.

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/helper/ -run TestSignedNumericalValidate`
Expected: FAIL, `undefined: SignedNumericalValidate`

- [ ] **Step 3: Write the implementation**

Add to `internal/helper/inputvalidation.go`, below `NumericalValidate`:

```go
// SignedNumericalValidate allows a leading minus, which NumericalValidate does
// not: the auction filter's minimum stat has to express penalties.
func SignedNumericalValidate(s string) error {
	matched, _ := regexp.MatchString(`^-?\d*$`, s)
	if !matched {
		return fmt.Errorf("%s", lokyn.L("Only numbers are allowed"))
	}
	return nil
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/helper/ -run TestSignedNumericalValidate`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/helper/inputvalidation.go internal/helper/inputvalidation_test.go
git commit -m "feat: signed numeric input validation"
```

---

### Task 2: Filter option type and filter building

The pure core of the filter widget: the selector value type, and the function turning raw control values into an `api.AuctionFilter`. Both are testable without a theme.

**Files:**
- Create: `widget/auctionfilter/option.go`
- Create: `widget/auctionfilter/filter.go`
- Test: `widget/auctionfilter/filter_test.go`

**Interfaces:**
- Consumes: `api.AuctionFilter`, `api.AuctionItemKind` from `core/data/api`.
- Produces:
  - `auctionfilter.Option{Code, Label string}` with `RenderValue() string`, satisfying `multivalueselector.Value`.
  - `buildFilter(kind, slot, weaponType, statCode, skillCode, minStat string) api.AuctionFilter` (package-private).

- [ ] **Step 1: Write the failing test**

Create `widget/auctionfilter/filter_test.go`:

```go
package auctionfilter

import (
	"farental/core/data/api"
	"testing"
)

func TestBuildFilterEmpty(t *testing.T) {
	got := buildFilter("", "", "", "", "", "")

	if got != (api.AuctionFilter{}) {
		t.Errorf("buildFilter with no values = %+v, want the zero filter", got)
	}
}

func TestBuildFilterAllSet(t *testing.T) {
	got := buildFilter("equipment", "hea", "sword", "str", "", "5")

	want := api.AuctionFilter{
		Kind:           api.AuctionItemKindEquipment,
		SlotCode:       "hea",
		WeaponTypeCode: "sword",
		StatCode:       "str",
		MinStat:        5,
		HasMinStat:     true,
	}

	if got != want {
		t.Errorf("buildFilter = %+v, want %+v", got, want)
	}
}

// The server ignores a minimum with no stat to refine, so the widget must not
// claim one is set.
func TestBuildFilterMinStatNeedsAStat(t *testing.T) {
	got := buildFilter("", "", "", "", "", "5")

	if got.HasMinStat {
		t.Errorf("HasMinStat = true with no stat or skill, want false")
	}
}

func TestBuildFilterNegativeMinStat(t *testing.T) {
	got := buildFilter("", "", "", "", "swordsmanship", "-5")

	if got.SkillCode != "swordsmanship" {
		t.Errorf("SkillCode = %q, want %q", got.SkillCode, "swordsmanship")
	}

	if !got.HasMinStat || got.MinStat != -5 {
		t.Errorf("MinStat = %d (has: %v), want -5 (true)", got.MinStat, got.HasMinStat)
	}
}

func TestBuildFilterUnparsableMinStat(t *testing.T) {
	got := buildFilter("", "", "", "str", "", "-")

	if got.HasMinStat {
		t.Errorf("HasMinStat = true for an unparsable minimum, want false")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./widget/auctionfilter/`
Expected: FAIL, `undefined: buildFilter`

- [ ] **Step 3: Write the implementation**

Create `widget/auctionfilter/option.go`:

```go
package auctionfilter

// Option is one entry of a filter selector. Code is what the server expects,
// Label is what the player reads; an empty Code means the filter is off.
type Option struct {
	Code  string
	Label string
}

func (o Option) RenderValue() string {
	return o.Label
}
```

Create `widget/auctionfilter/filter.go`:

```go
package auctionfilter

import (
	"farental/core/data/api"
	"strconv"
)

// buildFilter assembles the query from raw control values. Kept free of the
// widgets so the rules the server enforces can be tested directly.
func buildFilter(kind, slot, weaponType, statCode, skillCode, minStat string) api.AuctionFilter {
	f := api.AuctionFilter{
		Kind:           api.AuctionItemKind(kind),
		SlotCode:       slot,
		WeaponTypeCode: weaponType,
		StatCode:       statCode,
		SkillCode:      skillCode,
	}

	// A minimum narrows a stat; alone it is a no-op the server drops.
	if statCode == "" && skillCode == "" {
		return f
	}

	value, err := strconv.Atoi(minStat)

	if err != nil {
		return f
	}

	f.MinStat = value
	f.HasMinStat = true

	return f
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./widget/auctionfilter/`
Expected: PASS (5 tests)

- [ ] **Step 5: Commit**

```bash
git add widget/auctionfilter/
git commit -m "feat: auction filter option type and query building"
```

---

### Task 3: Stat or skill picker dialog

Stats and skills both name a stat, and `/auction/all` rejects a request carrying both. One picker over both lists makes that rejection unreachable. The dialog embeds `screen/generic/selectionlist` exactly as `screen/dialog/itemselection` does, so it gets search for free.

**Files:**
- Create: `screen/dialog/statskillselection/statskillselection.go`
- Create: `screen/dialog/statskillselection/entrylistitem.go`

**Interfaces:**
- Consumes: `api.AuctionFilterOptionsResponse` (`Stats []api.StatResponse`, `Skills []api.SkillResponse`).
- Produces:
  - `statskillselection.EntryKind` with `EntryNone`, `EntryStat`, `EntrySkill`.
  - `statskillselection.Entry{Kind EntryKind, Code string, Label string}`.
  - `statskillselection.New(options *api.AuctionFilterOptionsResponse) *Screen`, returning the chosen `Entry` from `OnExit` (so it arrives as `orvyn.DialogExitMsg.Param`), or `nil` when cancelled.
  - `statskillselection.Constructor` — `widgetlist.ItemConstructor[Entry]`.

- [ ] **Step 1: Write the list item**

Create `screen/dialog/statskillselection/entrylistitem.go`:

```go
package statskillselection

import (
	"fmt"

	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type listItem struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data Entry
}

func Constructor(data Entry) widgetlist.ListItem[Entry] {
	w := new(listItem)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.UpdateData(data)

	w.OnBlur()

	return w
}

func (w *listItem) Resize(size orvyn.Size) {
	size.Height = 3

	w.BaseWidget.Resize(size)
}

func (w *listItem) UpdateData(data Entry) {
	w.data = data
}

func (w *listItem) GetData() Entry {
	return w.data
}

func (w *listItem) FilterValue() string {
	return w.data.Label
}

func (w *listItem) Render() string {
	contentSize := w.GetContentSize()

	label := w.data.Label

	switch w.data.Kind {
	case EntryStat:
		label = fmt.Sprintf("%s %s", label,
			orvyn.GetTheme().Style(theme.DimTextStyleID).Render("("+lokyn.L("stat")+")"))
	case EntrySkill:
		label = fmt.Sprintf("%s %s", label,
			orvyn.GetTheme().Style(theme.DimTextStyleID).Render("("+lokyn.L("skill")+")"))
	}

	return w.GetStyle().Width(contentSize.Width).
		Height(contentSize.Height).
		Render(label)
}
```

- [ ] **Step 2: Write the dialog**

Create `screen/dialog/statskillselection/statskillselection.go`:

```go
package statskillselection

import (
	"farental/core/data/api"
	"farental/internal/keybind"
	"farental/screen/generic/selectionlist"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type EntryKind int

const (
	// EntryNone clears the stat filter.
	EntryNone EntryKind = iota
	EntryStat
	EntrySkill
)

// Entry is one pickable value. A skill stands for its primordial stat, which
// is why stats and skills share a single control.
type Entry struct {
	Kind  EntryKind
	Code  string
	Label string
}

type Screen struct {
	selectionlist.Screen[Entry]

	options *api.AuctionFilterOptionsResponse

	submitted bool
}

var _ orvyn.Screen = (*Screen)(nil)

func New(options *api.AuctionFilterOptionsResponse) *Screen {
	s := new(Screen)

	s.options = options

	s.Screen = selectionlist.New(lokyn.L("Stat or skill"),
		Constructor, s.loadData, s.submit)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	s.submitted = false

	cmd := s.Screen.OnEnter(i)

	s.SetTitle(lokyn.L("Stat or skill"))

	bubblehelp.SwitchContext(keybind.ContextFilterSelectionListBasic)

	return cmd
}

func (s *Screen) OnExit() any {
	if s.submitted {
		return s.GetSelectedItem()
	}

	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if m, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(m, keybind.Enter):
			if s.GetFilteringState() != widgetlist.Filtering {
				s.submitted = true
				return orvyn.CloseDialog()
			}

		case key.Matches(m, keybind.Esc):
			if s.GetFilteringState() == widgetlist.Unfiltered {
				return orvyn.CloseDialog()
			}
		}
	}

	return s.Screen.Update(msg)
}

func (s *Screen) loadData() {
	entries := []Entry{{Kind: EntryNone, Label: lokyn.L("Any stat or skill")}}

	if s.options != nil {
		for _, stat := range s.options.Stats {
			entries = append(entries, Entry{
				Kind:  EntryStat,
				Code:  stat.Code,
				Label: stat.Name,
			})
		}

		for _, skill := range s.options.Skills {
			entries = append(entries, Entry{
				Kind:  EntrySkill,
				Code:  skill.Code,
				Label: skill.Name,
			})
		}
	}

	s.SetItems(entries)
}

// submit satisfies selectionlist's callback; the dialog closes from Update,
// which is where it knows it is a dialog and not a screen.
func (s *Screen) submit() bool {
	return true
}
```

- [ ] **Step 3: Verify it builds**

Run: `gofmt -l . && go build ./... && go vet ./...`
Expected: no output from `gofmt`, no build or vet errors

- [ ] **Step 4: Commit**

```bash
git add screen/dialog/statskillselection/
git commit -m "feat: stat or skill picker dialog for the auction filter"
```

---

### Task 4: Filter widget

Fills in `widget/auctionfilter`, which today is an empty struct. Composite focusable with its own `FocusManager` walking five controls with ↑/↓ — Tab stays with the screen. Modelled on `widget/auctionstartform`.

**Files:**
- Modify: `widget/auctionfilter/auctionfilter.go`

**Interfaces:**
- Consumes: `buildFilter`, `Option` (Task 2), `statskillselection.New`/`Entry`/`EntryNone` (Task 3), `helper.SignedNumericalValidate` (Task 1).
- Produces:
  - `auctionfilter.New() *Widget`
  - `(*Widget).Init()`, `(*Widget).Reset()`
  - `(*Widget).SetOptions(options *api.AuctionFilterOptionsResponse)`
  - `(*Widget).GetFilter() api.AuctionFilter`
  - `auctionfilter.AppliedMsg` (`int`) and `auctionfilter.AppliedCmd() tea.Msg`

- [ ] **Step 1: Write the widget**

Replace the contents of `widget/auctionfilter/auctionfilter.go`:

```go
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

// AppliedMsg tells the screen the player asked for the current filter to be
// run. The widget never queries anything itself.
type AppliedMsg int

func AppliedCmd() tea.Msg {
	return AppliedMsg(1)
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

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
	w.lblKind.SetValue(lokyn.L("Kind"))
	w.lblSlot.SetValue(lokyn.L("Equipment slot"))
	w.lblWeapon.SetValue(lokyn.L("Weapon type"))
	w.lblStatSkill.SetValue(lokyn.L("Stat or skill"))
	w.lblMinStat.SetValue(lokyn.L("Minimum"))

	w.tiMinStat.Placeholder = lokyn.L("Minimum value")

	return nil
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case orvyn.DialogExitMsg:
		if msg.DialogID == dialogIDStatSkill {
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

	w.Reset()
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

// Reset returns every control to "no filter".
func (w *Widget) Reset() {
	w.mvsKind.SetSelected(0)
	w.mvsSlot.SetSelected(0)
	w.mvsWeapon.SetSelected(0)

	w.statCode = ""
	w.skillCode = ""

	w.btStatSkill.SetLabel(lokyn.L("Any stat or skill"))
	w.tiMinStat.SetValue("")

	w.focusManager.FocusFirst()
}

// GetFilter reads the controls into the query the request layer sends.
func (w *Widget) GetFilter() api.AuctionFilter {
	return buildFilter(
		w.mvsKind.GetSelectedValue().Code,
		w.mvsSlot.GetSelectedValue().Code,
		w.mvsWeapon.GetSelectedValue().Code,
		w.statCode,
		w.skillCode,
		w.tiMinStat.Value(),
	)
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
```

- [ ] **Step 2: Verify it builds and the Task 2 tests still pass**

Run: `gofmt -l . && go build ./... && go vet ./... && go test ./widget/auctionfilter/`
Expected: no gofmt/build/vet output, tests PASS

- [ ] **Step 3: Commit**

```bash
git add widget/auctionfilter/
git commit -m "feat: auction house filter widget"
```

---

### Task 5: Auction list row

One row per listing: item, quantity, current bid (marked when the player holds it), buy price or `—`, time left, seller. `EndsIn` is pure and gets the only test.

**Files:**
- Create: `widget/auctionlistitem/endsin.go`
- Create: `widget/auctionlistitem/auctionlistitem.go`
- Test: `widget/auctionlistitem/endsin_test.go`

**Interfaces:**
- Consumes: `api.AuctionResponse`, `context.CharacterInfo` (`*api.CharacterInfoResponse`, may be nil).
- Produces:
  - `auctionlistitem.EndsIn(end, now time.Time) string`
  - `auctionlistitem.Constructor` — `widgetlist.ItemConstructor[api.AuctionResponse]`

- [ ] **Step 1: Write the failing test**

Create `widget/auctionlistitem/endsin_test.go`:

```go
package auctionlistitem

import (
	"testing"
	"time"
)

func TestEndsIn(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		end  time.Time
		want string
	}{
		{"hours and minutes", now.Add(4*time.Hour + 2*time.Minute), "4h02"},
		{"whole hours", now.Add(12 * time.Hour), "12h00"},
		{"more than a day", now.Add(27 * time.Hour), "1d 3h"},
		{"under an hour", now.Add(45 * time.Minute), "45m"},
		{"already over", now.Add(-time.Minute), "ended"},
	}

	for _, c := range cases {
		if got := EndsIn(c.end, now); got != c.want {
			t.Errorf("%s: EndsIn = %q, want %q", c.name, got, c.want)
		}
	}
}
```

`"ended"` goes through `lokyn.L`, which returns the source string when no translation is loaded — that is what the test runs against.

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./widget/auctionlistitem/`
Expected: FAIL, `undefined: EndsIn`

- [ ] **Step 3: Write the implementation**

Create `widget/auctionlistitem/endsin.go`:

```go
package auctionlistitem

import (
	"fmt"
	"time"

	"github.com/halsten-dev/lokyn"
)

// EndsIn renders how long a listing has left. now is a parameter so the
// formatting can be tested without waiting for a clock.
func EndsIn(end, now time.Time) string {
	remaining := end.Sub(now)

	if remaining <= 0 {
		return lokyn.L("ended")
	}

	hours := int(remaining.Hours())

	if hours >= 24 {
		return fmt.Sprintf("%dd %dh", hours/24, hours%24)
	}

	if hours >= 1 {
		return fmt.Sprintf("%dh%02d", hours, int(remaining.Minutes())%60)
	}

	return fmt.Sprintf("%dm", int(remaining.Minutes()))
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./widget/auctionlistitem/`
Expected: PASS

- [ ] **Step 5: Write the row widget**

Create `widget/auctionlistitem/auctionlistitem.go`:

```go
package auctionlistitem

import (
	"farental/art"
	"farental/core/data/api"
	"farental/internal/context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data api.AuctionResponse

	ownBid bool
}

func Constructor(data api.AuctionResponse) widgetlist.ListItem[api.AuctionResponse] {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.UpdateData(data)

	w.OnBlur()

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 3

	w.BaseWidget.Resize(size)
}

func (w *Widget) UpdateData(data api.AuctionResponse) {
	w.data = data
	w.ownBid = false

	info := context.CharacterInfo

	if info == nil || data.CurrentBidderName == "" {
		return
	}

	// The server sends readable names, not IDs, so this is the only way to tell
	// the player they already hold the bid.
	w.ownBid = data.CurrentBidderName == fmt.Sprintf("%s %s", info.FirstName, info.LastName)
}

func (w *Widget) GetData() api.AuctionResponse {
	return w.data
}

func (w *Widget) FilterValue() string {
	return w.data.Item.Name
}

func (w *Widget) Render() string {
	contentSize := w.GetContentSize()

	t := orvyn.GetTheme()
	ns := lipgloss.NewStyle()

	left := fmt.Sprintf("%s x%d", w.data.Item.Name, w.data.Quantity)

	bid := fmt.Sprintf("%s %d%c", lokyn.L("bid"), w.data.CurrentBid, art.CharGrynars)

	if w.ownBid {
		bid = fmt.Sprintf("%s (%s)", bid, lokyn.L("you"))
	}

	buy := fmt.Sprintf("%s —", lokyn.L("buy"))

	if w.data.DirectBuyPrice > 0 {
		buy = fmt.Sprintf("%s %d%c", lokyn.L("buy"), w.data.DirectBuyPrice, art.CharGrynars)
	}

	right := strings.Join([]string{
		bid,
		buy,
		EndsIn(w.data.EndTimestamp, time.Now()),
		w.data.SellerName,
	}, "  ")

	width1, width2 := orvyn.DivideSizeFull(contentSize.Width)

	return w.GetStyle().Width(contentSize.Width).
		Height(contentSize.Height).
		Render(lipgloss.JoinHorizontal(lipgloss.Top,
			ns.Width(width1).
				AlignHorizontal(lipgloss.Left).
				Render(left),
			ns.Width(width2).
				AlignHorizontal(lipgloss.Right).
				Render(t.Style(theme.DimTextStyleID).Render(right))))
}
```

- [ ] **Step 6: Verify**

Run: `gofmt -l . && go build ./... && go vet ./... && go test ./widget/auctionlistitem/`
Expected: clean, tests PASS

- [ ] **Step 7: Commit**

```bash
git add widget/auctionlistitem/
git commit -m "feat: auction list row widget"
```

---

### Task 6: Auction inspect dialog

`i` on a row. Reuses `widget/iteminspect` for the item and adds the auction's own facts, which no other screen shows.

**Files:**
- Create: `screen/dialog/auctioninspect/auctioninspect.go`

**Interfaces:**
- Consumes: `api.AuctionResponse`, `iteminspect.New()` / `(*iteminspect.Widget).UpdateData(*api.ItemResponse)`, `EndsIn` (Task 5).
- Produces: `auctioninspect.New(auction api.AuctionResponse) *Screen`.

- [ ] **Step 1: Write the dialog**

Create `screen/dialog/auctioninspect/auctioninspect.go`:

```go
package auctioninspect

import (
	"farental/art"
	"farental/core/data/api"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/widget/auctionlistitem"
	"farental/widget/help"
	"farental/widget/iteminspect"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	inspector *iteminspect.Widget

	details *orvyn.SimpleRenderable

	help *help.Widget

	layout *layout.CenterLayout

	auction api.AuctionResponse
}

var _ orvyn.Screen = (*Screen)(nil)

func New(auction api.AuctionResponse) *Screen {
	s := new(Screen)

	s.auction = auction

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable(auction.Item.Name)
	s.title.Style = t.Style(theme.TitleStyleID)

	s.inspector = iteminspect.New()

	s.details = orvyn.NewSimpleRenderable("")

	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(10, t.Size(ftheme.LayoutWidthSizeID),
			orvyn.NewSize(10, 4),
			s.title,
			s.inspector,
			s.details,
			s.help),
	)

	return s
}

func (s *Screen) OnEnter(any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextBackAndQuit)

	s.title.SetValue(s.auction.Item.Name)
	s.inspector.UpdateData(&s.auction.Item)
	s.details.SetValue(s.detailLines())

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if k, ok := orvyn.GetKeyMsg(msg); ok {
		if key.Matches(k, keybind.Esc) {
			return orvyn.CloseDialog()
		}
	}

	return nil
}

func (s *Screen) detailLines() string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s: %d\n", lokyn.L("Quantity"), s.auction.Quantity)
	fmt.Fprintf(&b, "%s: %d%c\n", lokyn.L("Current bid"), s.auction.CurrentBid, art.CharGrynars)

	if s.auction.DirectBuyPrice > 0 {
		fmt.Fprintf(&b, "%s: %d%c\n", lokyn.L("Direct buy"), s.auction.DirectBuyPrice, art.CharGrynars)
	}

	bidder := s.auction.CurrentBidderName

	if bidder == "" {
		bidder = lokyn.L("nobody")
	}

	fmt.Fprintf(&b, "%s: %s\n", lokyn.L("Current bidder"), bidder)
	fmt.Fprintf(&b, "%s: %s\n", lokyn.L("Seller"), s.auction.SellerName)
	fmt.Fprintf(&b, "%s: %s", lokyn.L("Ends in"), auctionlistitem.EndsIn(s.auction.EndTimestamp, time.Now()))

	return b.String()
}
```

- [ ] **Step 2: Verify**

Run: `gofmt -l . && go build ./... && go vet ./...`
Expected: clean

- [ ] **Step 3: Commit**

```bash
git add screen/dialog/auctioninspect/
git commit -m "feat: auction inspect dialog"
```

---

### Task 7: Bid dialog

Enter on a row. A numeric input prefilled one grynar above the current bid, showing the player's money so an unaffordable bid is obvious before the server says so.

**Files:**
- Create: `screen/dialog/auctionbid/auctionbid.go`

**Interfaces:**
- Consumes: `api.AuctionResponse`, `helper.NumericalValidate`.
- Produces: `auctionbid.New(auction api.AuctionResponse, money int) *Screen`, returning the entered bid as `int` from `OnExit` (0 when cancelled or invalid).

- [ ] **Step 1: Write the dialog**

Create `screen/dialog/auctionbid/auctionbid.go`:

```go
package auctionbid

import (
	"farental/art"
	"farental/core/data/api"
	"farental/internal/helper"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/widget/help"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/label"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/textinput"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	lblInfo *label.Widget
	tiBid   *textinput.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	layout *layout.CenterLayout

	auction api.AuctionResponse
	money   int

	submitted bool
}

var _ orvyn.Screen = (*Screen)(nil)

func New(auction api.AuctionResponse, money int) *Screen {
	s := new(Screen)

	s.auction = auction
	s.money = money

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable("")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.lblInfo = label.New("")

	s.tiBid = textinput.New()
	s.tiBid.Validate = helper.NumericalValidate
	s.tiBid.Prompt = string(art.CharGrynars)

	s.statusMessage = statusmessage.New()

	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(10, t.Size(ftheme.LayoutWidthSizeID),
			orvyn.NewSize(10, 4),
			s.title,
			s.lblInfo,
			s.tiBid,
			s.statusMessage,
			s.help),
	)

	return s
}

func (s *Screen) OnEnter(any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextNavEnterEsc)

	s.submitted = false

	s.title.SetValue(fmt.Sprintf("%s x%d", s.auction.Item.Name, s.auction.Quantity))

	s.lblInfo.SetValue(fmt.Sprintf("%s %d%c | %s %d%c",
		lokyn.L("Current bid"), s.auction.CurrentBid, art.CharGrynars,
		lokyn.L("You have"), s.money, art.CharGrynars))

	s.tiBid.SetValue(strconv.Itoa(s.auction.CurrentBid + 1))
	s.tiBid.OnFocus()

	s.statusMessage.Reset()

	return nil
}

// OnExit returns the bid to place, 0 when the dialog was cancelled.
func (s *Screen) OnExit() any {
	if !s.submitted {
		return 0
	}

	bid, err := strconv.Atoi(s.tiBid.Value())

	if err != nil {
		return 0
	}

	return bid
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if k, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(k, keybind.Esc):
			return orvyn.CloseDialog()

		case key.Matches(k, keybind.Enter):
			bid, err := strconv.Atoi(s.tiBid.Value())

			if err != nil || bid <= s.auction.CurrentBid {
				s.statusMessage.SetMessage(
					fmt.Sprintf(lokyn.L("Your bid must be above %d%c"),
						s.auction.CurrentBid, art.CharGrynars),
					statusmessage.WarningMessage)

				return nil
			}

			s.submitted = true

			return orvyn.CloseDialog()
		}
	}

	return s.tiBid.Update(msg)
}
```

`keybind.ContextNavEnterEsc` already exists (`internal/keybind/context.go:51`) and is what `screen/auctionhouse/sell` uses.

- [ ] **Step 2: Verify**

Run: `gofmt -l . && go build ./... && go vet ./...`
Expected: clean

- [ ] **Step 3: Commit**

```bash
git add screen/dialog/auctionbid/
git commit -m "feat: auction bid dialog"
```

---

### Task 8: Buy screen — browsing

The screen itself: layout, focus, filter options, apply, load more. Actions land in Task 9, so this task ends with a screen that browses and is reachable from the menu.

**Files:**
- Modify: `screen/auctionhouse/buy/buy.go` (currently a stub returning nil everywhere)
- Modify: `internal/keybind/context.go` (add the help context)
- Modify: `main.go:179-180` area (register the screen)
- Modify: `screen/auctionhouse/menu/menu.go:109-111` (wire the button and the `b` key)

**Interfaces:**
- Consumes: `auctionfilter.New/SetOptions/GetFilter/Reset/Init/AppliedMsg`, `auctionlistitem.Constructor`, `request.AuctionGetAll`, `request.AuctionGetFilterOptions`.
- Produces: `buy.New() *Screen`, `keybind.ContextAuctionBuy`.

- [ ] **Step 1: Add the help context**

In `internal/keybind/context.go`, add to the context const block:

```go
	ContextAuctionBuy                          bubblehelp.KeymapContext = "auctionBuy"
```

And register the keymap next to the other `bubblehelp.RegisterContext` calls:

```go
	auctionBuyKeymap := bubblehelp.NewKeymap(3)
	auctionBuyKeymap.Style = mainHelpStyle
	auctionBuyKeymap.NewKeyBinding(Up, false)
	auctionBuyKeymap.NewKeyBinding(Down, false)
	auctionBuyKeymap.NewKeyBinding(Tab, true)
	auctionBuyKeymap.SetHelpDesc(Tab, lokyn.L("filters / list"))
	auctionBuyKeymap.NewKeyBinding(Enter, true)
	auctionBuyKeymap.SetHelpDesc(Enter, lokyn.L("apply filter / bid"))
	auctionBuyKeymap.NewKeyBinding(Space, false)
	auctionBuyKeymap.NewKeyBinding(BKey, true)
	auctionBuyKeymap.SetHelpDesc(BKey, lokyn.L("buy now"))
	auctionBuyKeymap.NewKeyBinding(IKey, true)
	auctionBuyKeymap.SetHelpDesc(IKey, lokyn.L("information"))
	auctionBuyKeymap.NewKeyBinding(MKey, true)
	auctionBuyKeymap.SetHelpDesc(MKey, lokyn.L("load more"))
	auctionBuyKeymap.NewKeyBinding(Esc, true)
	auctionBuyKeymap.NewKeyBinding(Quit, true)
	auctionBuyKeymap.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextAuctionBuy, auctionBuyKeymap)
```

- [ ] **Step 2: Write the screen**

Replace the contents of `screen/auctionhouse/buy/buy.go`:

```go
package buy

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
	"farental/widget/auctionfilter"
	"farental/widget/auctionlistitem"
	"farental/widget/characterinfo"
	"farental/widget/help"
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

// listIndex is the auction list's position in the screen's focus manager.
const listIndex = 1

type Screen struct {
	title *orvyn.SimpleRenderable

	characterInfo *characterinfo.Widget

	auctionFilter *auctionfilter.Widget
	auctionList   *widgetlist.Widget[api.AuctionResponse]

	statusMessage *statusmessage.Widget
	help          *help.Widget

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout

	options *api.AuctionFilterOptionsResponse

	// filter is the applied query, not the one being edited: load more has to
	// page the query that produced the rows on screen.
	filter api.AuctionFilter
	page   int
	total  int64

	money int
}

var _ orvyn.Screen = (*Screen)(nil)

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable("buy")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.characterInfo = characterinfo.New()

	s.auctionFilter = auctionfilter.New()

	s.auctionList = widgetlist.New(auctionlistitem.Constructor)
	s.auctionList.SetFilterable(false)
	s.auctionList.SetMinSize(orvyn.NewSize(20, 6))

	s.statusMessage = statusmessage.New()
	s.help = help.New()

	s.focusManager = orvyn.NewFocusManager()

	panels := []layout.FixedRatioRenderable{
		layout.NewFixedRatioRenderable(0.25, s.auctionFilter),
		layout.NewFixedRatioRenderable(0.75, s.auctionList),
	}

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(
			orvyn.NewSize(10, 4),
			3,
			s.title,
			s.characterInfo,
			orvyn.VGap,
			layout.NewHBoxFixedRatioLayout(0, 1, 1, panels...),
			s.statusMessage,
			s.help),
	)

	return s
}

func (s *Screen) OnEnter(any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextAuctionBuy)

	s.title.SetValue(lokyn.L("Auction house"))

	s.focusManager.SetWidgets([]orvyn.Focusable{s.auctionFilter, s.auctionList})
	s.focusManager.FocusFirst()

	s.statusMessage.Reset()

	s.auctionFilter.Init()

	s.updateCharacterInfo()
	s.loadFilterOptions()
	s.applyFilter()

	return nil
}

func (s *Screen) OnExit() any {
	s.focusManager.BlurCurrent()

	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if k, ok := orvyn.GetKeyMsg(msg); ok {
		switch {
		case key.Matches(k, keybind.Esc):
			return orvyn.SwitchToPreviousScreen()

		case key.Matches(k, keybind.MKey):
			if s.listFocused() {
				s.statusMessage.Reset()
				s.loadMore()

				return nil
			}
		}
	}

	switch msg.(type) {
	case auctionfilter.AppliedMsg:
		s.statusMessage.Reset()
		s.applyFilter()

		return nil
	}

	return s.focusManager.Update(msg)
}

func (s *Screen) listFocused() bool {
	return s.focusManager.TabIndex() == listIndex
}

// loadFilterOptions fetches the vocabulary on every entry: the labels are
// localized server-side, so a language change has to be picked up.
func (s *Screen) loadFilterOptions() {
	options, err := helper.Fetch[api.AuctionFilterOptionsResponse](
		request.AuctionGetFilterOptions())

	if err != nil {
		// Browsing unfiltered still works, so this degrades rather than blocks.
		s.statusMessage.SetError(err)
		return
	}

	s.options = options
	s.auctionFilter.SetOptions(options)
}

// applyFilter runs the filter currently in the panel from page 1, dropping
// whatever was accumulated.
func (s *Screen) applyFilter() {
	s.filter = s.auctionFilter.GetFilter()
	s.page = 1

	resp, err := helper.Fetch[api.AuctionListResponse](
		request.AuctionGetAll(s.page, s.filter))

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.total = resp.Total

	s.auctionList.SetItems(resp.Auctions)
	s.auctionList.FocusFirst()

	s.reportCount()
}

// loadMore appends the next server page, keeping the applied filter.
func (s *Screen) loadMore() {
	if int64(s.auctionList.Length()) >= s.total {
		s.statusMessage.SetMessage(lokyn.L("Everything is already loaded"),
			statusmessage.InformationMessage)

		return
	}

	resp, err := helper.Fetch[api.AuctionListResponse](
		request.AuctionGetAll(s.page+1, s.filter))

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.page++
	s.total = resp.Total

	s.auctionList.SetItems(append(s.auctionList.GetItems(), resp.Auctions...))

	s.reportCount()
}

func (s *Screen) reportCount() {
	if s.auctionList.Length() == 0 {
		s.statusMessage.SetMessage(lokyn.L("No auction matches these filters"),
			statusmessage.InformationMessage)

		return
	}

	s.statusMessage.SetMessage(fmt.Sprintf("%d/%d %s",
		s.auctionList.Length(), s.total, lokyn.L("auction(s)")),
		statusmessage.InformationMessage)
}

func (s *Screen) updateCharacterInfo() {
	characterInfo, currency, err := context.RefreshCharacterInfo(true)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.money = currency

	unreadMail := context.RefreshHaveUnreadMail()

	data := characterinfo.ConvertCharacterInfoResponseToData(
		characterInfo, currency, unreadMail)
	s.characterInfo.UpdateData(data)
}
```

The auction list is built with `SetFilterable(false)`: widgetlist's text filter
keeps stale indices across `SetItems` and panics on render, and the server-side
filter makes it redundant here.

- [ ] **Step 3: Register the screen**

In `main.go`, next to the existing auction house registrations (~line 180):

```go
	orvyn.RegisterScreen(screen.IDAuctionHouseBuy, buy.New())
```

Add the import `"farental/screen/auctionhouse/buy"` beside `"farental/screen/auctionhouse/sell"`.

- [ ] **Step 4: Wire the menu**

In `screen/auctionhouse/menu/menu.go`, make the Buy button switch screens:

```go
func (s *Screen) btBuyOnClicked() tea.Cmd {
	return orvyn.SwitchScreen(screen.IDAuctionHouseBuy)
}
```

And add the shortcut beside the existing `SKey` case in `Update`:

```go
		case key.Matches(k, keybind.BKey):
			return orvyn.SwitchScreen(screen.IDAuctionHouseBuy)
```

- [ ] **Step 5: Verify**

Run: `gofmt -l . && go build ./... && go vet ./... && go test ./...`
Expected: clean, all tests PASS

- [ ] **Step 6: Commit**

```bash
git add screen/auctionhouse/buy/ screen/auctionhouse/menu/menu.go internal/keybind/context.go main.go
git commit -m "feat: auction house buy screen browsing"
```

---

### Task 9: Buy screen — inspect, bid, buy

Adds the three actions to the screen built in Task 8.

**Files:**
- Modify: `screen/auctionhouse/buy/buy.go`

**Interfaces:**
- Consumes: `auctioninspect.New` (Task 6), `auctionbid.New` (Task 7), `popup.NewYesNo`, `request.AuctionMakeBid`, `request.AuctionDirectBuy`.
- Produces: nothing for later tasks.

- [ ] **Step 1: Add the dialog IDs and key handling**

In `screen/auctionhouse/buy/buy.go`, add the constants below `listIndex`:

```go
const (
	dialogIDInspect    orvyn.ScreenID = "auctionInspect"
	dialogIDBid        orvyn.ScreenID = "auctionBid"
	dialogIDBuyConfirm orvyn.ScreenID = "auctionBuyConfirm"
)
```

Extend the key switch in `Update` (after the `MKey` case):

```go
		case key.Matches(k, keybind.IKey):
			if s.listFocused() && s.auctionList.Length() > 0 {
				return orvyn.OpenDialog(dialogIDInspect,
					auctioninspect.New(s.auctionList.GetSelectedItem()), nil)
			}

		case key.Matches(k, keybind.BKey):
			if s.listFocused() && s.auctionList.Length() > 0 {
				return s.openBuyConfirm()
			}

		case key.Matches(k, keybind.Enter):
			if s.listFocused() && s.auctionList.Length() > 0 {
				return orvyn.OpenDialog(dialogIDBid,
					auctionbid.New(s.auctionList.GetSelectedItem(), s.money), nil)
			}
		}
```

Extend the message switch:

```go
	switch msg := msg.(type) {
	case auctionfilter.AppliedMsg:
		s.statusMessage.Reset()
		s.applyFilter()

		return nil

	case orvyn.DialogExitMsg:
		switch msg.DialogID {
		case dialogIDBid:
			bid, ok := msg.Param.(int)

			if ok && bid > 0 {
				s.placeBid(bid)
			}

			return nil

		case dialogIDBuyConfirm:
			answer, ok := msg.Param.(uint)

			if ok && answer == 1 {
				s.directBuy()
			}

			return nil
		}
	}
```

The unhandled dialog IDs must fall through to `focusManager.Update`: the
stat/skill picker belongs to the filter widget, and swallowing its exit here
would drop the player's pick silently. Return only inside the cases this screen
owns.

- [ ] **Step 2: Add the action helpers**

Append to `screen/auctionhouse/buy/buy.go`:

```go
func (s *Screen) openBuyConfirm() tea.Cmd {
	auction := s.auctionList.GetSelectedItem()

	if auction.DirectBuyPrice <= 0 {
		s.statusMessage.SetMessage(lokyn.L("This auction has no direct buy price"),
			statusmessage.WarningMessage)

		return nil
	}

	return orvyn.OpenDialog(dialogIDBuyConfirm, popup.NewYesNo(
		fmt.Sprintf(lokyn.L("Buy %s x%d for %d%c ?"),
			auction.Item.Name, auction.Quantity,
			auction.DirectBuyPrice, art.CharGrynars)), nil)
}

func (s *Screen) placeBid(bid int) {
	auction := s.auctionList.GetSelectedItem()

	_, err := helper.SendRequest(request.AuctionMakeBid(api.AuctionBidBody{
		ID:  auction.ID,
		Bid: bid,
	}))

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.afterAction(lokyn.L("Bid placed"))
}

func (s *Screen) directBuy() {
	auction := s.auctionList.GetSelectedItem()

	_, err := helper.SendRequest(request.AuctionDirectBuy(api.IDBody{ID: auction.ID}))

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.afterAction(lokyn.L("Item bought, check your mail"))
}

// afterAction reloads from page 1: a sold listing is gone and every later page
// shifts under it, so refetching the accumulated pages would show duplicates
// or holes.
func (s *Screen) afterAction(message string) {
	s.updateCharacterInfo()
	s.applyFilter()

	s.statusMessage.SetMessage(message, statusmessage.SuccessMessage)
}
```

Add the imports this needs: `farental/art`, `farental/screen/dialog/auctionbid`, `farental/screen/dialog/auctioninspect`, `farental/screen/dialog/popup`.

`api.IDBody` is `struct{ ID uint }` (`core/data/api/body.go:3`).

- [ ] **Step 3: Verify**

Run: `gofmt -l . && go build ./... && go vet ./... && go test ./...`
Expected: clean, all tests PASS

- [ ] **Step 4: Commit**

```bash
git add screen/auctionhouse/buy/buy.go
git commit -m "feat: auction house bidding and direct buy"
```

---

## Final verification

- [ ] Run the full suite from `src/`: `gofmt -l . && go build ./... && go vet ./... && go test ./...`
- [ ] State plainly in the report that the screen itself was not exercised live — it needs a TTY and a running backend. Do not claim otherwise.
