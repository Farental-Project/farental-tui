# Auction House Local Search Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a free-text field to the auction house filter panel that narrows the already-fetched auction list by fuzzy match on the item name.

**Architecture:** The list widget (`orvyn/widget/widgetlist`) already owns fuzzy matching, the filtered-index mapping and the paginator maths behind it, but only the `/` key can reach it. Tasks 1-3 fix two latent bugs on that path and open a public door (`SetFilter`, `VisibleLength`). Tasks 4-6 add the field to `auctionfilter`, wire it through the buy screen as applied state alongside the server query, and stand down the screen's single-letter keybinds while the field has focus.

**Tech Stack:** Go, Bubble Tea, orvyn (TUI widget library), lokyn (i18n), `github.com/sahilm/fuzzy`.

## Global Constraints

- **Two repositories.** Tasks 1-3 are in `/home/halsten/Dev/Libs/Go/orvyn` (branch `main`, currently clean). Tasks 4-6 are in `/home/halsten/Dev/Farental/farental-tui` (branch `auction-local-search`). Each task's commit goes in its own repo.
- `farental-tui/src/go.mod:57` has `replace github.com/halsten-dev/orvyn => /home/halsten/Dev/Libs/Go/orvyn`. orvyn edits take effect immediately; **no version bump or tag is needed.**
- All `go test ./...` commands for farental run from `/home/halsten/Dev/Farental/farental-tui/src`, not the repo root.
- Both repos are green at the start of this plan. Keep them green.
- Comments explain *why*, never *what*. Match the density and voice of the surrounding code — see `buy.go:435-439` for the house style.
- Player-visible strings go through `lokyn.L(...)`.
- Widget unit tests must bootstrap with `orvyn.Init()`, `lokyn.Init()`, `lokyn.SetLanguage("en")` in that order. Without the last two, `lokyn.L` dereferences a nil localizer and the test panics.
- Run `gofmt -l .` and `go vet ./...` before each commit.

---

### Task 1: Record the applied needle, and make an empty needle clear

The needle currently lives in `tiFilter`'s value, which `SetItem`/`AppendItem` read back to re-apply. A filter driven from outside the widget never touches that field, so the needle has to become state of its own before anything else can be built on it.

`filter("")` also falls through its empty case: it clears, then immediately re-filters with `""` and stamps `FilterApplied` back over the cleared state. `fuzzy.Find("", data)` returns **zero** matches, so the list goes blank.

**Files:**
- Modify: `/home/halsten/Dev/Libs/Go/orvyn/widget/widgetlist/widgetlist.go`
- Test: `/home/halsten/Dev/Libs/Go/orvyn/widget/widgetlist/widgetlist_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces: unexported field `filterNeedle string` on `Widget[T]`, holding the needle of the applied filter and `""` when unfiltered. Tasks 2 and 3 read it.

- [ ] **Step 1: Write the failing tests**

Append to `widget/widgetlist/widgetlist_test.go`:

```go
// The applied needle has to be state of its own. It used to be read back out
// of tiFilter, which a filter driven from outside the widget never touches.
func TestFilterRecordsTheNeedle(t *testing.T) {
	w := newTestList(t, 20, orvyn.NewSize(20, 10))

	w.filter("item 1")

	if w.filterNeedle != "item 1" {
		t.Errorf("filterNeedle = %q, want %q", w.filterNeedle, "item 1")
	}
}

// An empty needle is not a filter. fuzzy.Find returns no match for one, so
// falling through the empty case leaves the list showing nothing at all.
func TestFilterEmptyNeedleClears(t *testing.T) {
	w := newTestList(t, 20, orvyn.NewSize(20, 10))

	w.filter("item 1")

	if w.filterState != FilterApplied {
		t.Fatalf("test setup: filterState = %v, want FilterApplied", w.filterState)
	}

	w.filter("")

	if w.filterState != Unfiltered {
		t.Errorf("filterState = %v, want Unfiltered", w.filterState)
	}

	if len(w.filteredListItems) != 0 {
		t.Errorf("filteredListItems = %d entries, want 0", len(w.filteredListItems))
	}

	if w.filterNeedle != "" {
		t.Errorf("filterNeedle = %q, want empty", w.filterNeedle)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd /home/halsten/Dev/Libs/Go/orvyn && go test ./widget/widgetlist/ -run 'TestFilterRecordsTheNeedle|TestFilterEmptyNeedleClears' -v`

Expected: build failure, `w.filterNeedle undefined (type *Widget[string] has no field or method filterNeedle)`.

- [ ] **Step 3: Add the field**

In `widget/widgetlist/widgetlist.go`, in the `Widget[T]` struct, put it directly under `filterState`:

```go
	filterable                bool
	blockCursorMovingCallback bool
	filterState               FilterState

	// filterNeedle is the needle of the applied filter. It is kept here rather
	// than read back out of tiFilter: a filter driven from outside the widget
	// (SetFilter) leaves that field untouched, and every re-apply path would
	// then re-run with an empty needle.
	filterNeedle string
```

- [ ] **Step 4: Rewrite `filter` and `clearFilter`**

Replace `filter`:

```go
func (w *Widget[T]) filter(s string) {
	// An empty needle is not a filter. clearFilter already restores the
	// unfiltered state; falling through would stamp FilterApplied back over it
	// with whatever the matcher returns for "" - which, for the fuzzy matcher,
	// is nothing at all, blanking the list.
	if s == "" {
		w.clearFilter()
		w.FocusFirst()

		return
	}

	w.tiFilter.OnBlur()

	w.filterNeedle = s
	w.filteredListItems = w.Filter(&w.listItems, s)

	w.filterState = FilterApplied

	w.paginatorUpdate()

	w.FocusFirst()
}
```

In `clearFilter`, add the needle reset next to the other state it clears:

```go
func (w *Widget[T]) clearFilter() {
	w.tiFilter.SetValue("")
	w.tiFilter.OnBlur()

	w.filterNeedle = ""
	w.filteredListItems = make(FilteredItems, 0)
```

- [ ] **Step 5: Point the two re-apply paths at the needle**

In `SetItem`, replace `w.filter(w.tiFilter.Value())` with `w.filter(w.filterNeedle)`.

In `AppendItem`, replace `w.filter(w.tiFilter.Value())` with `w.filter(w.filterNeedle)`.

Leave `AppendItem`'s opening `w.clearFilter()` alone. It makes the `FilterApplied` branch below it unreachable, but that predates this work and is not in scope.

- [ ] **Step 6: Run the whole widgetlist suite**

Run: `cd /home/halsten/Dev/Libs/Go/orvyn && go test ./widget/widgetlist/ -v`

Expected: PASS, including the pre-existing `TestFilteredListShrinks` and `TestSetItemsShrinksList`.

- [ ] **Step 7: Run the full suite, format and vet**

Run: `cd /home/halsten/Dev/Libs/Go/orvyn && go test ./... && gofmt -l . && go vet ./...`

Expected: tests pass, `gofmt -l` prints nothing.

- [ ] **Step 8: Commit**

```bash
cd /home/halsten/Dev/Libs/Go/orvyn
git add widget/widgetlist/widgetlist.go widget/widgetlist/widgetlist_test.go
git commit -m "fix: don't blank the list when filtering on an empty needle"
```

---

### Task 2: Re-apply the filter on SetItems, and leave it alone on Esc

`SetItem` and `AppendItem` re-run the filter after they edit the list; `SetItems` does not. It replaces the whole slice, so `filteredListItems` is left holding positions into the previous one — the same class of stale index that caused the selection-list panics.

`Update` also clears the filter on Esc regardless of `filterable`. That was unreachable while `filter` was private, but Task 3 makes it reachable, and it would leave the panel's field showing text the list no longer applies.

**Files:**
- Modify: `/home/halsten/Dev/Libs/Go/orvyn/widget/widgetlist/widgetlist.go`
- Test: `/home/halsten/Dev/Libs/Go/orvyn/widget/widgetlist/widgetlist_test.go`

**Interfaces:**
- Consumes: `filterNeedle` from Task 1.
- Produces: test helper `newNamedList(t *testing.T, items []string, size orvyn.Size) *Widget[string]`, used again in Task 3. `SetItems` now preserves an applied filter.

- [ ] **Step 1: Write the failing tests**

Append to `widget/widgetlist/widgetlist_test.go`. The named-item helper exists because `newTestList`'s `"item 0".."item 19"` all fuzzy-match each other, which makes match counts hard to assert on:

```go
func newNamedList(t *testing.T, items []string, size orvyn.Size) *Widget[string] {
	t.Helper()

	orvyn.Init()

	w := New(SimpleListItemConstructor)
	w.SetFilterable(false)
	w.Resize(size)
	w.SetItems(items)

	return w
}

// SetItems replaces the slice the applied filter indexes into. Without a
// re-run, filteredListItems still points into the old one.
func TestSetItemsReappliesTheFilter(t *testing.T) {
	w := newNamedList(t, []string{"iron sword", "steel sword", "iron helm"},
		orvyn.NewSize(20, 10))

	w.filter("sword")

	if len(w.filteredListItems) != 2 {
		t.Fatalf("test setup: %d matches, want 2", len(w.filteredListItems))
	}

	w.SetItems([]string{"oak staff", "iron sword", "leather cap"})

	if len(w.filteredListItems) != 1 {
		t.Fatalf("filteredListItems = %d entries, want 1", len(w.filteredListItems))
	}

	if got := w.filteredListItems[0].Index; got != 1 {
		t.Errorf("filtered index = %d, want 1", got)
	}

	checkBounds(t, w)

	w.Render()
}

// Esc belongs to the built-in filter UI. A list filtered from elsewhere in the
// screen has to keep its filter, or the field driving it is left showing text
// the list no longer applies.
func TestEscKeepsTheFilterWhenNotFilterable(t *testing.T) {
	w := newNamedList(t, []string{"iron sword", "steel sword", "iron helm"},
		orvyn.NewSize(20, 10))

	w.filter("sword")

	w.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if w.filterState != FilterApplied {
		t.Errorf("filterState after esc = %v, want FilterApplied", w.filterState)
	}
}

// The same key still clears a list that owns its own filter field.
func TestEscClearsTheFilterWhenFilterable(t *testing.T) {
	w := newNamedList(t, []string{"iron sword", "steel sword", "iron helm"},
		orvyn.NewSize(20, 10))
	w.SetFilterable(true)

	w.filter("sword")

	w.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if w.filterState != Unfiltered {
		t.Errorf("filterState after esc = %v, want Unfiltered", w.filterState)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd /home/halsten/Dev/Libs/Go/orvyn && go test ./widget/widgetlist/ -run 'TestSetItemsReappliesTheFilter|TestEscKeepsTheFilterWhenNotFilterable|TestEscClearsTheFilterWhenFilterable' -v`

Expected: `TestSetItemsReappliesTheFilter` fails with `filteredListItems = 2 entries, want 1`, `TestEscKeepsTheFilterWhenNotFilterable` fails with `filterState after esc = unfiltered, want FilterApplied`. `TestEscClearsTheFilterWhenFilterable` already passes.

- [ ] **Step 3: Re-apply the filter in `SetItems`**

In `SetItems`, between `w.focusManager.SetWidgets(focusableList)` and the `paginatorUpdate` call:

```go
	w.focusManager.SetWidgets(focusableList)

	// An applied filter indexes into the slice that was just replaced. Re-run
	// it before anything reads filteredListItems. Not through filter(): that
	// resets the cursor to the first match, and SetItems deliberately keeps
	// the global index so the paginator maths below can clamp it.
	if w.filterState == FilterApplied {
		w.filteredListItems = w.Filter(&w.listItems, w.filterNeedle)
	}

	// paginatorUpdate clamps the global index against the new list, so focus
	// after it: focusing first would use the index of the previous, possibly
	// longer list, and Focus ignores an out-of-range index.
	w.paginatorUpdate()
```

- [ ] **Step 4: Gate the Esc clear on `filterable`**

In `Update`, in the `!isInputting` branch:

```go
			case key.Matches(msg, w.keybinds.clearFilter):
				// Esc belongs to the built-in filter UI. A list whose filter is
				// driven from outside (SetFilter) keeps it: the field driving
				// it lives elsewhere and would be left out of sync.
				if w.filterable && w.filterState == FilterApplied {
					w.clearFilter()

					w.FocusFirst()

					return nil
				}
```

- [ ] **Step 5: Run the whole widgetlist suite**

Run: `cd /home/halsten/Dev/Libs/Go/orvyn && go test ./widget/widgetlist/ -v`

Expected: PASS.

- [ ] **Step 6: Run the full suite, format and vet**

Run: `cd /home/halsten/Dev/Libs/Go/orvyn && go test ./... && gofmt -l . && go vet ./...`

Expected: tests pass, `gofmt -l` prints nothing.

- [ ] **Step 7: Commit**

```bash
cd /home/halsten/Dev/Libs/Go/orvyn
git add widget/widgetlist/widgetlist.go widget/widgetlist/widgetlist_test.go
git commit -m "fix: keep an applied filter valid across SetItems"
```

---

### Task 3: Open the filter to callers outside the widget

**Files:**
- Modify: `/home/halsten/Dev/Libs/Go/orvyn/widget/widgetlist/widgetlist.go`
- Test: `/home/halsten/Dev/Libs/Go/orvyn/widget/widgetlist/widgetlist_test.go`

**Interfaces:**
- Consumes: `filterNeedle` (Task 1), `newNamedList` (Task 2).
- Produces:
  - `func (w *Widget[T]) SetFilter(s string)` — applies `s`; `""` clears.
  - `func (w *Widget[T]) VisibleLength() int` — filtered count while filtered, full count otherwise. Task 5 uses both.

- [ ] **Step 1: Write the failing tests**

Append to `widget/widgetlist/widgetlist_test.go`:

```go
func TestSetFilterNarrowsTheList(t *testing.T) {
	w := newNamedList(t, []string{"iron sword", "steel shield", "iron helm"},
		orvyn.NewSize(20, 10))

	w.SetFilter("sword")

	if got := w.VisibleLength(); got != 1 {
		t.Fatalf("VisibleLength = %d, want 1", got)
	}

	// The rows are narrowed, not dropped: the screen still has to be able to
	// report how many it has loaded.
	if got := w.Length(); got != 3 {
		t.Errorf("Length = %d, want 3", got)
	}

	if got := w.GetSelectedItem(); got != "iron sword" {
		t.Errorf("GetSelectedItem = %q, want %q", got, "iron sword")
	}
}

func TestSetFilterEmptyRestoresTheFullList(t *testing.T) {
	w := newNamedList(t, []string{"iron sword", "steel shield", "iron helm"},
		orvyn.NewSize(20, 10))

	w.SetFilter("sword")
	w.SetFilter("")

	if w.filterState != Unfiltered {
		t.Errorf("filterState = %v, want Unfiltered", w.filterState)
	}

	if got := w.VisibleLength(); got != 3 {
		t.Errorf("VisibleLength = %d, want 3", got)
	}
}

// A needle that matches nothing leaves the widget with no cursor at all, which
// Render and the paginator both have to survive.
func TestSetFilterWithNoMatch(t *testing.T) {
	w := newNamedList(t, []string{"iron sword", "steel shield", "iron helm"},
		orvyn.NewSize(20, 10))

	w.SetFilter("zzzz")

	if got := w.VisibleLength(); got != 0 {
		t.Errorf("VisibleLength = %d, want 0", got)
	}

	checkBounds(t, w)

	w.Render()
}

// A list that still shows the built-in field must keep it in step with the
// filter a caller applied.
func TestSetFilterMirrorsIntoTheBuiltInField(t *testing.T) {
	w := newNamedList(t, []string{"iron sword", "steel shield", "iron helm"},
		orvyn.NewSize(20, 10))
	w.SetFilterable(true)

	w.SetFilter("sword")

	if got := w.tiFilter.Value(); got != "sword" {
		t.Errorf("tiFilter value = %q, want %q", got, "sword")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd /home/halsten/Dev/Libs/Go/orvyn && go test ./widget/widgetlist/ -run TestSetFilter -v`

Expected: build failure, `w.SetFilter undefined` and `w.VisibleLength undefined`.

- [ ] **Step 3: Add the two methods**

In `widget/widgetlist/widgetlist.go`, next to `SetFilterable` and `SetFilterPlaceholder`:

```go
// SetFilter applies s from outside the widget, for lists whose filter field
// lives elsewhere in the screen (SetFilterable(false)). An empty s clears the
// filter. The value is mirrored into the built-in field so a list that still
// shows it does not fall out of step.
func (w *Widget[T]) SetFilter(s string) {
	w.tiFilter.SetValue(s)
	w.filter(s)
}
```

And next to `Length`:

```go
// VisibleLength returns how many items the list is currently showing: the
// filtered count while a filter is applied, the full count otherwise. Length
// always reports the full count, which is what a caller wants for "how much
// have I loaded" and never for "is there a row under the cursor".
func (w *Widget[T]) VisibleLength() int {
	if w.filterState == FilterApplied {
		return len(w.filteredListItems)
	}

	return len(w.listItems)
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd /home/halsten/Dev/Libs/Go/orvyn && go test ./widget/widgetlist/ -v`

Expected: PASS.

- [ ] **Step 5: Run the full suite, format and vet**

Run: `cd /home/halsten/Dev/Libs/Go/orvyn && go test ./... && gofmt -l . && go vet ./...`

Expected: tests pass, `gofmt -l` prints nothing.

- [ ] **Step 6: Commit**

```bash
cd /home/halsten/Dev/Libs/Go/orvyn
git add widget/widgetlist/widgetlist.go widget/widgetlist/widgetlist_test.go
git commit -m "feat: let a caller drive the list filter"
```

---

### Task 4: Add the search field to the filter panel

The field goes first in the panel — first in the layout, first in the focus order — so `w.focusManager.TabIndex() == 0` means the search box has the panel's cursor.

**Files:**
- Modify: `/home/halsten/Dev/Farental/farental-tui/src/widget/auctionfilter/auctionfilter.go`
- Modify: `/home/halsten/Dev/Farental/farental-tui/src/translations/en.json`, `de.json`, `fr.json`
- Test: `/home/halsten/Dev/Farental/farental-tui/src/widget/auctionfilter/auctionfilter_test.go` (create)

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces, on `*auctionfilter.Widget`:
  - `func (w *Widget) GetSearch() string` — the local search text, `""` when empty.
  - `func (w *Widget) SearchFocused() bool` — true only when the panel is focused *and* its cursor is on the search box.
  - `Reset()` additionally clears the search box.
  - `GetFilter()` is unchanged and must stay unaware of the search text.

  Tasks 5 and 6 consume all four.

- [ ] **Step 1: Write the failing tests**

Create `src/widget/auctionfilter/auctionfilter_test.go`:

```go
package auctionfilter

import (
	"farental/core/data/api"
	"testing"

	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
)

// New reaches lokyn through Reset, which dereferences a localizer that only
// SetLanguage creates.
func newTestWidget(t *testing.T) *Widget {
	t.Helper()

	orvyn.Init()
	lokyn.Init()
	lokyn.SetLanguage("en")

	w := New()
	w.Init()

	return w
}

// The search box narrows rows that are already on screen, so its text must
// never reach the query the request layer sends.
func TestSearchNeverReachesTheQuery(t *testing.T) {
	w := newTestWidget(t)

	w.tiSearch.SetValue("iron sword")

	if got := w.GetSearch(); got != "iron sword" {
		t.Errorf("GetSearch = %q, want %q", got, "iron sword")
	}

	if got := w.GetFilter(); got != (api.AuctionFilter{}) {
		t.Errorf("GetFilter = %+v, want the zero filter", got)
	}
}

func TestResetClearsTheSearch(t *testing.T) {
	w := newTestWidget(t)

	w.tiSearch.SetValue("iron sword")

	w.Reset()

	if got := w.GetSearch(); got != "" {
		t.Errorf("GetSearch after Reset = %q, want empty", got)
	}
}

// The screen stands its single-letter keybinds down on this, so it has to be
// false whenever the panel does not hold the screen's focus - the panel keeps
// its own cursor where it was while the auction list is being used.
func TestSearchFocusedNeedsThePanelToHoldFocus(t *testing.T) {
	w := newTestWidget(t)

	screenFocus := orvyn.NewFocusManager()
	screenFocus.SetWidgets([]orvyn.Focusable{w})
	screenFocus.FocusFirst()

	if !w.SearchFocused() {
		t.Errorf("SearchFocused = false with the panel focused on its first control, want true")
	}

	screenFocus.BlurCurrent()

	if w.SearchFocused() {
		t.Errorf("SearchFocused = true with the panel blurred, want false")
	}
}

// Moving off the box is what gives the screen its letter keys back.
func TestSearchFocusedFollowsThePanelCursor(t *testing.T) {
	w := newTestWidget(t)

	screenFocus := orvyn.NewFocusManager()
	screenFocus.SetWidgets([]orvyn.Focusable{w})
	screenFocus.FocusFirst()

	w.focusManager.NextFocus()

	if w.SearchFocused() {
		t.Errorf("SearchFocused = true with the panel cursor moved off the box, want false")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go test ./widget/auctionfilter/ -v`

Expected: build failure, `w.tiSearch undefined`, `w.GetSearch undefined`, `w.SearchFocused undefined`.

- [ ] **Step 3: Add the field to the widget**

In `src/widget/auctionfilter/auctionfilter.go`.

Add the constant under `dialogIDStatSkill`:

```go
// searchIndex is the search box's position in the panel's focus manager. The
// screen asks whether the cursor sits there before it claims a letter key.
const searchIndex = 0
```

Add to the `Widget` struct, above `lblScope`:

```go
	lblSearch *label.Widget
	tiSearch  *textinput.Widget

	lblScope *label.Widget
	mvsScope *multivalueselector.Widget[Option]
```

In `New`, before `w.lblScope = label.New("")`:

```go
	w.lblSearch = label.New("")

	// No Validate: this one takes any text the player types.
	w.tiSearch = textinput.New()

	w.lblScope = label.New("")
```

In `New`, make the search box the first focus stop:

```go
	w.focusManager = orvyn.NewFocusManager()
	w.focusManager.Add(w.tiSearch)
	w.focusManager.Add(w.mvsScope)
```

In `New`, make it the first two rows of the layout:

```go
	w.layout = layout.NewMaxWidthVBoxLayout(0,
		w.lblSearch,
		w.tiSearch,
		w.lblScope,
		w.mvsScope,
```

In `Init`, with the other labels:

```go
	w.lblSearch.SetValue(lokyn.L("Search"))
	w.lblScope.SetValue(lokyn.L("Show"))
```

and with the other placeholder:

```go
	w.tiSearch.Placeholder = lokyn.L("Filter loaded auctions")
	w.tiMinStat.Placeholder = lokyn.L("Minimum value")
```

In `Reset`, with the other cleared control:

```go
	w.tiSearch.SetValue("")
	w.btStatSkill.SetLabel(lokyn.L("Any stat or skill"))
	w.tiMinStat.SetValue("")
```

- [ ] **Step 4: Add the two accessors**

In `src/widget/auctionfilter/auctionfilter.go`, directly after `GetFilter`:

```go
// GetSearch reads the local search box. It is deliberately absent from
// GetFilter: the text narrows the rows already fetched and never reaches the
// server.
func (w *Widget) GetSearch() string {
	return w.tiSearch.Value()
}

// SearchFocused reports whether the free-text box holds the panel's cursor.
// The screen stands down the single-letter keybinds on it, which would
// otherwise swallow what the player types.
//
// The panel's own focus is not enough: its cursor stays where it was while the
// auction list is being used, so the screen's focus has to be on the panel
// too.
func (w *Widget) SearchFocused() bool {
	return w.IsFocused() && w.focusManager.TabIndex() == searchIndex
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go test ./widget/auctionfilter/ -v`

Expected: PASS.

- [ ] **Step 6: Add the two strings to all three translation files**

The files are single-line JSON keyed by the English source string. Edit them with a script rather than by hand, from `/home/halsten/Dev/Farental/farental-tui/src`:

```bash
python3 - <<'PY'
import json

additions = {
    "en.json": {"Search": "Search",
                "Filter loaded auctions": "Filter loaded auctions"},
    "de.json": {"Search": "Suche",
                "Filter loaded auctions": "Geladene Auktionen filtern"},
    "fr.json": {"Search": "Recherche",
                "Filter loaded auctions": "Filtrer les enchères chargées"},
}

for name, entries in additions.items():
    path = "translations/" + name
    with open(path, encoding="utf-8") as f:
        data = json.load(f)
    data.update(entries)
    with open(path, "w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, separators=(",", ":"))
PY
```

- [ ] **Step 7: Verify the translation files still parse and gained both keys**

Run:

```bash
cd /home/halsten/Dev/Farental/farental-tui/src
python3 -c "
import json
for n in ('en','de','fr'):
    d = json.load(open('translations/%s.json' % n, encoding='utf-8'))
    print(n, len(d), repr(d['Search']), repr(d['Filter loaded auctions']))
"
```

Expected: three lines, each with 384 keys and both strings present.

- [ ] **Step 8: Run the full suite, format and vet**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go build ./... && go test ./... && gofmt -l . && go vet ./...`

Expected: builds, tests pass, `gofmt -l` prints nothing.

- [ ] **Step 9: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/widget/auctionfilter/auctionfilter.go \
        src/widget/auctionfilter/auctionfilter_test.go \
        src/translations/en.json src/translations/de.json src/translations/fr.json
git commit -m "feat: add a local search box to the auction filter panel"
```

---

### Task 5: Carry the search through the buy screen

The screen already distinguishes the *applied* query from the one being edited: `r`, and the reloads that follow a bid or a buy, re-run `s.filter` rather than whatever the panel currently holds. The search gets the same treatment, for the same reason.

**Files:**
- Modify: `/home/halsten/Dev/Farental/farental-tui/src/screen/auctionhouse/buy/buy.go`

**Interfaces:**
- Consumes: `SetFilter`, `VisibleLength`, `Length` (Task 3); `GetSearch` (Task 4).
- Produces: `reload` now takes `(filter api.AuctionFilter, search string, mayReport bool) bool`. Task 6 does not touch it.

- [ ] **Step 1: Add the applied-search field**

In the `Screen` struct, extend the existing comment block:

```go
	// filter is the applied query, not the one being edited: load more has to
	// page the query that produced the rows on screen.
	filter api.AuctionFilter
	// search is the applied local search, held for the same reason: a reload
	// must re-narrow with what the player committed, not an edit they have not
	// pressed Enter on.
	search string
	page   int
	total  int64
```

- [ ] **Step 2: Thread the search through `applyFilter` and `reload`**

Replace `applyFilter`'s body:

```go
func (s *Screen) applyFilter(mayReport bool) bool {
	return s.reload(s.auctionFilter.GetFilter(), s.auctionFilter.GetSearch(),
		mayReport)
}
```

Change `reload`'s signature and its committing block:

```go
func (s *Screen) reload(filter api.AuctionFilter, search string, mayReport bool) bool {
	resp, err := helper.Fetch[api.AuctionListResponse](
		request.AuctionGetAll(1, filter))

	if err != nil {
		s.statusMessage.SetError(err)
		return false
	}

	s.filter = filter
	s.search = search
	s.page = 1
	s.total = resp.Total

	s.auctionList.SetItems(resp.Auctions)
	s.auctionList.SetFilter(search)
	s.auctionList.FocusFirst()

	if mayReport {
		s.reportCount()
	}

	return true
}
```

`loadMore` needs no change: it goes through `SetItems`, which now re-applies the filter itself.

- [ ] **Step 3: Update the two other `reload` call sites**

In `update`, the `keybind.RKey` case:

```go
			case key.Matches(k, keybind.RKey):
				// Reloads using the applied filter (s.filter) and search
				// (s.search), not applyFilter: a player who edited the panel
				// without pressing Enter should not have that edit silently
				// committed by a refresh.
				s.statusMessage.Reset()
				s.reload(s.filter, s.search, true)

				return nil
```

In `afterAction`:

```go
	infoOK := s.updateCharacterInfo()
	filterOK := s.reload(s.filter, s.search, infoOK)
```

and extend that function's doc comment's first sentence:

```go
// afterAction reloads from page 1 using the applied filter (s.filter) and
// search (s.search), not the panel: a player who edited the controls without
// pressing Enter and then bid or bought should not have that reload silently
// pick up the unconfirmed edit. A sold listing is gone and every later page
// shifts under it, so refetching the accumulated pages would show duplicates
// or holes.
```

- [ ] **Step 4: Point the three "is there a row" guards at the visible count**

In `update`, `Length()` answers "how much have I loaded", which is not the same question once a search narrows the list. Change all three:

```go
			case key.Matches(k, keybind.IKey):
				if s.auctionList.VisibleLength() > 0 {
```

```go
			case key.Matches(k, keybind.BKey):
				if s.listFocused() && s.auctionList.VisibleLength() > 0 {
					return s.openBuyConfirm()
				}
```

```go
			case key.Matches(k, keybind.Enter):
				if s.listFocused() && s.auctionList.VisibleLength() > 0 {
```

Leave `loadMore`'s `int64(s.auctionList.Length()) >= s.total` as it is: that one really is asking how much has been loaded.

- [ ] **Step 5: Report the narrowed count**

Replace `reportCount`:

```go
func (s *Screen) reportCount() {
	if s.auctionList.VisibleLength() == 0 {
		s.statusMessage.SetMessage(lokyn.L("No auction matches these filters"),
			statusmessage.InformationMessage)

		return
	}

	// The loaded count stays in the line while a search narrows it: without it,
	// a short list reads as if there were nothing left for load more to fetch.
	if s.search != "" {
		s.statusMessage.SetMessage(fmt.Sprintf("%d/%d/%d %s",
			s.auctionList.VisibleLength(), s.auctionList.Length(), s.total,
			lokyn.L("auction(s)")),
			statusmessage.InformationMessage)

		return
	}

	s.statusMessage.SetMessage(fmt.Sprintf("%d/%d %s",
		s.auctionList.Length(), s.total, lokyn.L("auction(s)")),
		statusmessage.InformationMessage)
}
```

- [ ] **Step 6: Build, test, format and vet**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go build ./... && go test ./... && gofmt -l . && go vet ./...`

Expected: builds, tests pass, `gofmt -l` prints nothing.

- [ ] **Step 7: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/screen/auctionhouse/buy/buy.go
git commit -m "feat: narrow the auction list with the panel's search box"
```

---

### Task 6: Stop the screen's letter keys from eating what is typed

`buy.go` claims `m`, `r`, `b` and `i` before it delegates to the focus manager. Typing `iron` into the search box would fire inspect and reload. `m` and `b` are already gated on the auction list holding focus, so the panel never reaches them; `r` and `i` are not. `Ctrl+R` cannot be typed into a text field and needs nothing.

**Files:**
- Modify: `/home/halsten/Dev/Farental/farental-tui/src/screen/auctionhouse/buy/buy.go`

**Interfaces:**
- Consumes: `SearchFocused` (Task 4).
- Produces: nothing later tasks rely on. This is the last task.

- [ ] **Step 1: Guard the reload key**

In `update`, replace the whole `RKey` case. The guard is an `if` that falls out
of the switch rather than an early `break`, matching how `MKey` and `BKey`
already hand an unclaimed key on to the focus manager:

```go
			case key.Matches(k, keybind.RKey):
				if !s.searchFocused() {
					// Reloads using the applied filter (s.filter) and search
					// (s.search), not applyFilter: a player who edited the
					// panel without pressing Enter should not have that edit
					// silently committed by a refresh.
					s.statusMessage.Reset()
					s.reload(s.filter, s.search, true)

					return nil
				}
```

- [ ] **Step 2: Guard the inspect key and correct its comment**

The existing comment claims nothing in the panel can swallow the key, which the search box makes false. Replace the whole case:

```go
			case key.Matches(k, keybind.IKey):
				// Inspecting is read-only, so it stays available while the
				// filter panel holds focus - except over the search box, where
				// the letter is text the player is typing.
				if !s.searchFocused() && s.auctionList.VisibleLength() > 0 {
					return orvyn.OpenDialog(dialogIDInspect,
						auctioninspect.New(s.auctionList.GetSelectedItem()), nil)
				}
```

- [ ] **Step 3: Add the helper next to `listFocused`**

```go
// searchFocused reports whether the panel's free-text search box has the
// cursor. The screen claims several single letters before delegating, and over
// that box every one of them is text.
func (s *Screen) searchFocused() bool {
	return s.auctionFilter.SearchFocused()
}
```

- [ ] **Step 4: Take the two keys out of the help line while the box has focus**

In `updateFocusKeybinds`, after the `listFocused` block and before the Enter description, and extend the function's doc comment:

```go
	// The panel holds a free-text search box, and these two are single letters:
	// over that box they are text, not commands.
	searchFocused := s.searchFocused()

	bubblehelp.SetKeybindVisible(keybind.RKey, !searchFocused)
	bubblehelp.SetKeybindVisible(keybind.IKey, !searchFocused)
```

The doc comment's closing paragraph gains a sentence:

```go
// Space is deliberately absent: auctionfilter owns it, tying it to its own
// stat/skill button's focus.
//
// m and b need no search-box handling of their own: both are already gated on
// the auction list holding focus, so the panel never reaches them.
```

- [ ] **Step 5: Build, test, format and vet**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go build ./... && go test ./... && gofmt -l . && go vet ./...`

Expected: builds, tests pass, `gofmt -l` prints nothing.

- [ ] **Step 6: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/screen/auctionhouse/buy/buy.go
git commit -m "fix: don't let the screen's letter keys eat the search box"
```

---

## Manual verification

Automated tests cover the widgets; the interaction needs a real terminal and a
live backend. Run the client, open the auction house, and check:

1. The panel opens with `Search` above `Show`, and the cursor on the search box.
2. Type `iron` — the letters appear in the box, no reload fires, no inspect
   dialog opens, and the help line does not offer `r` or `i`.
3. Press Enter — the list refetches and shows only the fuzzy matches for
   `iron`, and the status line reads `shown/loaded/total`.
4. Press Down to leave the box — `r` and `i` come back to the help line and
   work again.
5. Press `m` on a narrowed list — a page is appended and the narrowing survives.
6. Press `r` — the list reloads and the narrowing survives.
7. Press `Ctrl+R` — the box is cleared along with everything else, and the full
   list comes back with the two-part count.
8. Type something that matches nothing and press Enter — the list is empty,
   the status line reads "No auction matches these filters", and `i`, `b` and
   Enter over the list do nothing rather than acting on a phantom row.
