# Auction Inspect Dialog Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the auction inspect dialog as two side-by-side bordered boxes — a new aligned auction-details widget on the left, the existing item inspector on the right.

**Architecture:** A new `widget/auctiondetails` package renders the auction facts as label/value rows with right-flushed values, inside the border every `orvyn` widget gets for free. The dialog swaps its flat `SimpleRenderable` for that widget and moves both boxes into an `HBoxFixedRatioLayout`. The name-comparison that decides whether the player holds the current bid is extracted from `auctionlistitem` so both widgets share it.

**Tech Stack:** Go 1.25, bubbletea, lipgloss, `github.com/halsten-dev/orvyn` (layout/widget framework), `github.com/halsten-dev/lokyn` (i18n), `github.com/charmbracelet/x/ansi`.

**Spec:** `docs/superpowers/specs/2026-08-04-auctioninspect-dialog-redesign-design.md`

## Global Constraints

- Working directory for every command is `/home/halsten/Dev/Farental/farental-tui/src` (the Go module root, `module farental`).
- Internal imports are module-absolute: `farental/widget/...`, `farental/core/data/api`, `farental/internal/...`. There are no relative imports.
- `farental/internal/theme` is always imported with the alias `ftheme`, because `github.com/halsten-dev/orvyn/theme` takes the bare name `theme`.
- All user-facing strings go through `lokyn.L("...")`. Any new key must be added to `translations/en.json`, `translations/fr.json` and `translations/de.json` in the same commit.
- `orvyn.NewBaseWidget()` reads the active theme, so any test that constructs a widget must call `orvyn.SetTheme(ftheme.NewFarentalDarkTheme())` in `TestMain` first.
- Only one `TestMain` may exist per package. `widget/auctionlistitem` already has one in `endsin_test.go` — do not add a second.
- `lipgloss` `Width(n)` wraps text rather than truncating it. Any content rendered into a fixed-height row must be cut with `ansi.Truncate` first.
- Do not modify `widget/iteminspect`. `screen/inventory` and `screen/shop` share it and are out of scope.
- Verification command for every task: `go build ./... && go vet ./... && go test ./... && gofmt -l .` — `gofmt -l .` must print nothing.

---

### Task 1: Extract `auctionlistitem.IsOwnBid`

The check lives inline in `UpdateData` today. `auctiondetails` needs the same
answer, so it moves to an exported function in the same package, next to the
already-exported `EndsIn`.

**Files:**
- Create: `widget/auctionlistitem/ownbid.go`
- Create: `widget/auctionlistitem/ownbid_test.go`
- Modify: `widget/auctionlistitem/auctionlistitem.go:47-60` (`UpdateData`), and its import block at lines 3-17

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: `func IsOwnBid(bidderName string) bool` in package `auctionlistitem`. Returns `false` when `context.CharacterInfo` is nil or when `bidderName` is empty.

- [ ] **Step 1: Write the failing test**

Create `widget/auctionlistitem/ownbid_test.go`. No `TestMain` here —
`endsin_test.go` already provides one for this package.

```go
package auctionlistitem

import (
	"farental/core/data/api"
	"farental/internal/context"
	"testing"
)

func TestIsOwnBid(t *testing.T) {
	cases := []struct {
		name   string
		info   *api.CharacterInfoResponse
		bidder string
		want   bool
	}{
		{
			name:   "no character loaded",
			info:   nil,
			bidder: "John Doe",
			want:   false,
		},
		{
			name:   "no bidder yet",
			info:   &api.CharacterInfoResponse{FirstName: "John", LastName: "Doe"},
			bidder: "",
			want:   false,
		},
		{
			name:   "someone else holds the bid",
			info:   &api.CharacterInfoResponse{FirstName: "John", LastName: "Doe"},
			bidder: "Jane Roe",
			want:   false,
		},
		{
			name:   "player holds the bid",
			info:   &api.CharacterInfoResponse{FirstName: "John", LastName: "Doe"},
			bidder: "John Doe",
			want:   true,
		},
		{
			name:   "partial name is not a match",
			info:   &api.CharacterInfoResponse{FirstName: "John", LastName: "Doe"},
			bidder: "John",
			want:   false,
		},
	}

	original := context.CharacterInfo
	defer func() { context.CharacterInfo = original }()

	for _, c := range cases {
		context.CharacterInfo = c.info

		if got := IsOwnBid(c.bidder); got != c.want {
			t.Errorf("%s: IsOwnBid(%q) = %v, want %v",
				c.name, c.bidder, got, c.want)
		}
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./widget/auctionlistitem/ -run TestIsOwnBid -v`
Expected: FAIL — `undefined: IsOwnBid`

- [ ] **Step 3: Write the implementation**

Create `widget/auctionlistitem/ownbid.go`:

```go
package auctionlistitem

import (
	"farental/internal/context"
	"fmt"
)

// IsOwnBid reports whether bidderName is the logged-in character. The server
// sends readable names, not IDs, so comparing the full name is the only way to
// tell the player they already hold the bid.
func IsOwnBid(bidderName string) bool {
	info := context.CharacterInfo

	if info == nil || bidderName == "" {
		return false
	}

	return bidderName == fmt.Sprintf("%s %s", info.FirstName, info.LastName)
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./widget/auctionlistitem/ -run TestIsOwnBid -v`
Expected: PASS

- [ ] **Step 5: Rewire `UpdateData` to the helper**

In `widget/auctionlistitem/auctionlistitem.go`, replace the whole body of
`UpdateData` (lines 47-60):

```go
func (w *Widget) UpdateData(data api.AuctionResponse) {
	w.data = data
	w.ownBid = IsOwnBid(data.CurrentBidderName)
}
```

`farental/internal/context` is now unused in this file — remove that one line
from the import block. Leave `fmt` alone; `Render` still uses it.

- [ ] **Step 6: Verify the package still builds and passes**

Run: `go build ./... && go vet ./... && go test ./... && gofmt -l .`
Expected: all pass, `gofmt -l .` prints nothing.

- [ ] **Step 7: Commit**

```bash
git add widget/auctionlistitem/ownbid.go widget/auctionlistitem/ownbid_test.go widget/auctionlistitem/auctionlistitem.go
git commit -m "refactor: extract IsOwnBid from the auction list item"
```

---

### Task 2: Build the auction detail rows

The pure part of the new widget: turning an `api.AuctionResponse` into the lines
the box will draw. Split from `Render` so it can be tested without a terminal.

**Files:**
- Create: `widget/auctiondetails/rows.go`
- Create: `widget/auctiondetails/rows_test.go`

**Interfaces:**
- Consumes: `auctionlistitem.IsOwnBid(string) bool` from Task 1, and the existing `auctionlistitem.EndsIn(end, now time.Time) string`.
- Produces:
  - `type row struct { label string; value string; highlight bool }` (unexported)
  - `func buildRows(auction *api.AuctionResponse, now time.Time) []row` (unexported)

  Both are used by Task 3 from inside the same package.

- [ ] **Step 1: Write the failing test**

Create `widget/auctiondetails/rows_test.go`:

```go
package auctiondetails

import (
	"farental/core/data/api"
	"farental/internal/context"
	"os"
	"testing"
	"time"

	"github.com/halsten-dev/lokyn"
)

func TestMain(m *testing.M) {
	lokyn.Init()
	lokyn.SetLanguage("en")

	os.Exit(m.Run())
}

// labels pulls the label column out of the rows so a test can assert the row
// set without caring about the values.
func labels(rows []row) []string {
	out := make([]string, 0, len(rows))

	for _, r := range rows {
		out = append(out, r.label)
	}

	return out
}

func find(t *testing.T, rows []row, label string) row {
	t.Helper()

	for _, r := range rows {
		if r.label == label {
			return r
		}
	}

	t.Fatalf("no row labelled %q in %v", label, labels(rows))

	return row{}
}

func TestBuildRowsWithoutDirectBuy(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)

	auction := api.AuctionResponse{
		CurrentBid:        120,
		DirectBuyPrice:    0,
		CurrentBidderName: "Jane Roe",
		SellerName:        "John Doe",
		EndTimestamp:      now.Add(2 * time.Hour),
	}

	rows := buildRows(&auction, now)

	want := []string{"Current bid", "Current bidder", "Seller", "Ends in"}

	if got := labels(rows); len(got) != len(want) {
		t.Fatalf("labels = %v, want %v", got, want)
	}

	for i, l := range labels(rows) {
		if l != want[i] {
			t.Errorf("row %d label = %q, want %q", i, l, want[i])
		}
	}
}

func TestBuildRowsWithDirectBuy(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)

	auction := api.AuctionResponse{
		CurrentBid:        120,
		DirectBuyPrice:    500,
		CurrentBidderName: "Jane Roe",
		SellerName:        "John Doe",
		EndTimestamp:      now.Add(2 * time.Hour),
	}

	rows := buildRows(&auction, now)

	directBuy := find(t, rows, "Direct buy")

	if directBuy.value != "500Ǥ" {
		t.Errorf("Direct buy value = %q, want %q", directBuy.value, "500Ǥ")
	}

	if !directBuy.highlight {
		t.Error("Direct buy should be highlighted as a money row")
	}
}

func TestBuildRowsEmptyBidderReadsNobody(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)

	auction := api.AuctionResponse{
		CurrentBid:   0,
		SellerName:   "John Doe",
		EndTimestamp: now.Add(2 * time.Hour),
	}

	rows := buildRows(&auction, now)

	if got := find(t, rows, "Current bidder").value; got != "nobody" {
		t.Errorf("Current bidder value = %q, want %q", got, "nobody")
	}
}

func TestBuildRowsMarksOwnBid(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)

	original := context.CharacterInfo
	defer func() { context.CharacterInfo = original }()

	context.CharacterInfo = &api.CharacterInfoResponse{
		FirstName: "John",
		LastName:  "Doe",
	}

	auction := api.AuctionResponse{
		CurrentBid:        120,
		CurrentBidderName: "John Doe",
		SellerName:        "Jane Roe",
		EndTimestamp:      now.Add(2 * time.Hour),
	}

	rows := buildRows(&auction, now)

	if got := find(t, rows, "Current bid").value; got != "120Ǥ (you)" {
		t.Errorf("Current bid value = %q, want %q", got, "120Ǥ (you)")
	}
}

func TestBuildRowsEndsInUsesSharedFormatter(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)

	auction := api.AuctionResponse{
		SellerName:   "John Doe",
		EndTimestamp: now.Add(4*time.Hour + 2*time.Minute),
	}

	rows := buildRows(&auction, now)

	if got := find(t, rows, "Ends in").value; got != "4h02" {
		t.Errorf("Ends in value = %q, want %q", got, "4h02")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./widget/auctiondetails/ -v`
Expected: FAIL — the package does not compile, `undefined: row`, `undefined: buildRows`

- [ ] **Step 3: Write the implementation**

Create `widget/auctiondetails/rows.go`:

```go
package auctiondetails

import (
	"farental/art"
	"farental/core/data/api"
	"farental/widget/auctionlistitem"
	"fmt"
	"time"

	"github.com/halsten-dev/lokyn"
)

// row is one label/value line of the box.
type row struct {
	label string
	value string

	// highlight marks the money rows. They are rendered with the highlight
	// style so the price is the first thing read.
	highlight bool
}

// buildRows turns an auction into the lines the widget draws. It is kept apart
// from Render so the row set can be tested without a terminal.
//
// Quantity is deliberately absent: the dialog carries it in its title as
// "<item name> x<quantity>".
func buildRows(auction *api.AuctionResponse, now time.Time) []row {
	bid := fmt.Sprintf("%d%c", auction.CurrentBid, art.CharGrynars)

	if auctionlistitem.IsOwnBid(auction.CurrentBidderName) {
		bid = fmt.Sprintf("%s (%s)", bid, lokyn.L("you"))
	}

	rows := []row{
		{label: lokyn.L("Current bid"), value: bid, highlight: true},
	}

	if auction.DirectBuyPrice > 0 {
		rows = append(rows, row{
			label:     lokyn.L("Direct buy"),
			value:     fmt.Sprintf("%d%c", auction.DirectBuyPrice, art.CharGrynars),
			highlight: true,
		})
	}

	bidder := auction.CurrentBidderName

	if bidder == "" {
		bidder = lokyn.L("nobody")
	}

	return append(rows,
		row{label: lokyn.L("Current bidder"), value: bidder},
		row{label: lokyn.L("Seller"), value: auction.SellerName},
		row{
			label: lokyn.L("Ends in"),
			value: auctionlistitem.EndsIn(auction.EndTimestamp, now),
		},
	)
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./widget/auctiondetails/ -v`
Expected: PASS, all five tests.

- [ ] **Step 5: Verify the whole module**

Run: `go build ./... && go vet ./... && go test ./... && gofmt -l .`
Expected: all pass, `gofmt -l .` prints nothing.

- [ ] **Step 6: Commit**

```bash
git add widget/auctiondetails/rows.go widget/auctiondetails/rows_test.go
git commit -m "feat: build the auction detail rows"
```

---

### Task 3: The `auctiondetails` widget

Wraps the rows in a bordered box with an underlined header and a right-flushed
value column.

**Files:**
- Create: `widget/auctiondetails/auctiondetails.go`
- Create: `widget/auctiondetails/auctiondetails_test.go`
- Modify: `widget/auctiondetails/rows_test.go` (its `TestMain` and import block — see Step 2)
- Modify: `translations/en.json`, `translations/fr.json`, `translations/de.json`

**Interfaces:**
- Consumes: `row` and `buildRows` from Task 2.
- Produces, in package `auctiondetails`:
  - `func New() *Widget`
  - `func (w *Widget) UpdateData(auction *api.AuctionResponse)`
  - `func (w *Widget) Render() string`
  - `func (w *Widget) GetMinSize() orvyn.Size`
  - `func (w *Widget) GetPreferredSize() orvyn.Size`

  Task 4 calls `New` and `UpdateData` only; the rest are the `orvyn.Widget`
  contract the layout calls.

- [ ] **Step 1: Add the one new translation key**

`"Auction"` is the only string this task introduces — every other label already
exists in all three catalogues. The files are single-line JSON objects; add the
key with a JSON-aware edit rather than by hand so the formatting survives:

```bash
python3 - <<'PY'
import json

values = {
    "translations/en.json": "Auction",
    "translations/fr.json": "Enchère",
    "translations/de.json": "Auktion",
}

for path, value in values.items():
    with open(path, encoding="utf-8") as f:
        data = json.load(f)

    data["Auction"] = value

    with open(path, "w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, separators=(",", ":"))
PY
```

Verify: `python3 -c "import json;[print(f, json.load(open(f))['Auction']) for f in ['translations/en.json','translations/fr.json','translations/de.json']]"`
Expected: `Auction`, `Enchère`, `Auktion`.

- [ ] **Step 2: Write the failing test**

Create `widget/auctiondetails/auctiondetails_test.go`. `TestMain` already lives
in `rows_test.go` from Task 2 — extend that one instead of adding a second, by
replacing its body with the version below, which also installs the theme that
`orvyn.NewBaseWidget()` needs.

Replace `TestMain` in `widget/auctiondetails/rows_test.go` with:

```go
func TestMain(m *testing.M) {
	lokyn.Init()
	lokyn.SetLanguage("en")

	orvyn.SetTheme(ftheme.NewFarentalDarkTheme())

	os.Exit(m.Run())
}
```

and add these two imports to that file's import block:

```go
	ftheme "farental/internal/theme"

	"github.com/halsten-dev/orvyn"
```

Then create `widget/auctiondetails/auctiondetails_test.go`:

```go
package auctiondetails

import (
	"farental/core/data/api"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
)

func testAuction(directBuy int) api.AuctionResponse {
	return api.AuctionResponse{
		CurrentBid:        120,
		DirectBuyPrice:    directBuy,
		CurrentBidderName: "Jane Roe",
		SellerName:        "John Doe",
		EndTimestamp:      time.Now().Add(2 * time.Hour),
	}
}

// The box must never render taller than the height it reports, or the dialog
// layout that trusts that number overflows.
func TestRenderedHeightMatchesReportedHeight(t *testing.T) {
	cases := []struct {
		name      string
		directBuy int
	}{
		{"without direct buy", 0},
		{"with direct buy", 500},
	}

	for _, c := range cases {
		w := New()

		auction := testAuction(c.directBuy)
		w.UpdateData(&auction)

		height := w.GetMinSize().Height

		if got := w.GetPreferredSize().Height; got != height {
			t.Errorf("%s: preferred height = %d, min height = %d; they must "+
				"match so the layout treats the box as fixed height",
				c.name, got, height)
		}

		w.Resize(orvyn.NewSize(41, height))

		if got := lipgloss.Height(w.Render()); got != height {
			t.Errorf("%s: rendered height = %d, reported %d",
				c.name, got, height)
		}
	}
}

// A value too wide for its column must be cut, not wrapped: a wrapped row would
// push the box past its reported height.
func TestLongValueDoesNotWrap(t *testing.T) {
	w := New()

	auction := testAuction(0)
	auction.SellerName = strings.Repeat("Wenceslas", 8)
	w.UpdateData(&auction)

	height := w.GetMinSize().Height

	w.Resize(orvyn.NewSize(41, height))

	if got := lipgloss.Height(w.Render()); got != height {
		t.Errorf("rendered height = %d, reported %d; a long value wrapped "+
			"instead of being truncated", got, height)
	}
}

func TestRenderShowsLabelsAndValues(t *testing.T) {
	w := New()

	auction := testAuction(500)
	w.UpdateData(&auction)

	w.Resize(orvyn.NewSize(41, w.GetMinSize().Height))

	view := w.Render()

	for _, want := range []string{"Auction", "Current bid", "Direct buy", "John Doe"} {
		if !strings.Contains(view, want) {
			t.Errorf("rendered box does not contain %q:\n%s", want, view)
		}
	}
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./widget/auctiondetails/ -v`
Expected: FAIL — `undefined: New`

- [ ] **Step 4: Write the implementation**

Create `widget/auctiondetails/auctiondetails.go`:

```go
package auctiondetails

import (
	"farental/core/data/api"
	ftheme "farental/internal/theme"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
)

// labelColumnRatio is the share of the box width given to the labels; the rest
// holds the right-flushed values. 0.55 is sized off the longest translated
// label rather than the English one - "Current bidder" is 14 characters but the
// French "Enchérisseur actuel" is 19.
const labelColumnRatio = 0.55

// Widget renders the auction facts as a bordered box of aligned label/value
// rows. It is a peer of iteminspect: the border comes from the default widget
// style, so the two boxes match without any extra styling.
type Widget struct {
	orvyn.BaseWidget

	rows []row
}

var _ orvyn.Widget = (*Widget)(nil)

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	return w
}

func (w *Widget) UpdateData(auction *api.AuctionResponse) {
	w.rows = buildRows(auction, time.Now())
}

func (w *Widget) Render() string {
	var b strings.Builder

	t := orvyn.GetTheme()
	contentSize := w.GetContentSize()

	b.WriteString(headerStyle().Width(contentSize.Width).
		Render(lokyn.L("Auction")))

	labelWidth := int(float64(contentSize.Width) * labelColumnRatio)
	valueWidth := max(contentSize.Width-labelWidth, 0)

	labelStyle := t.Style(theme.DimTextStyleID)
	valueStyle := t.Style(theme.NormalTextStyleID)
	moneyStyle := t.Style(theme.HighlightTextStyleID)

	ns := lipgloss.NewStyle()

	for _, r := range w.rows {
		vs := valueStyle

		if r.highlight {
			vs = moneyStyle
		}

		// Width(n) wraps rather than truncates, and a wrapped row would push the
		// box past the height GetMinSize reports. Cut both columns to their
		// budget before styling so neither can wrap.
		label := ansi.Truncate(r.label, labelWidth, "")
		value := ansi.Truncate(r.value, valueWidth, "")

		b.WriteString("\n")
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
			ns.Width(labelWidth).Render(labelStyle.Render(label)),
			ns.Width(valueWidth).AlignHorizontal(lipgloss.Right).
				Render(vs.Render(value))))
	}

	return w.GetStyle().
		Width(contentSize.Width).
		Height(contentSize.Height).
		Render(b.String())
}

// headerStyle is the box title treatment: the same underlined style iteminspect
// uses for its Stats / Equip conditions / Effects sections.
func headerStyle() lipgloss.Style {
	return orvyn.GetTheme().Style(ftheme.DimUnderlinedTextStyleID)
}

// GetMinSize and GetPreferredSize report the same height, which makes the
// layout treat the box as fixed height and hand it exactly its rows. Width
// stays at 1 so the widget never drives the layout width, matching
// iteminspect.GetMinSize.
func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(1, w.height())
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(1, w.height())
}

// height is what the box needs: the underlined header, one line per row, and
// the frame. The header height is measured rather than assumed so a theme that
// drops the underline does not leave a dead line behind.
func (w *Widget) height() int {
	header := lipgloss.Height(headerStyle().Render("X"))

	return header + len(w.rows) + w.GetStyle().GetVerticalFrameSize()
}
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./widget/auctiondetails/ -v`
Expected: PASS — the three tests from this task plus the five from Task 2.

- [ ] **Step 6: Verify the whole module**

Run: `go build ./... && go vet ./... && go test ./... && gofmt -l .`
Expected: all pass, `gofmt -l .` prints nothing.

- [ ] **Step 7: Commit**

```bash
git add widget/auctiondetails translations/en.json translations/fr.json translations/de.json
git commit -m "feat: add the auction details widget"
```

---

### Task 4: Rewire the dialog side by side

Swap the flat detail text for the new widget and put the two boxes in a
horizontal row.

**Files:**
- Modify: `screen/dialog/auctioninspect/auctioninspect.go` (whole file)

**Interfaces:**
- Consumes: `auctiondetails.New()` and `(*auctiondetails.Widget).UpdateData(*api.AuctionResponse)` from Task 3.
- Produces: nothing new. `auctioninspect.New(auction api.AuctionResponse) *Screen` keeps its signature, so `screen/auctionhouse/buy/buy.go:163` needs no change.

- [ ] **Step 1: Replace the file**

Rewrite `screen/dialog/auctioninspect/auctioninspect.go` in full. `detailLines`
is deleted along with the `art`, `strings`, `time`, `lokyn` and
`auctionlistitem` imports it needed; `fmt` stays for the title.

```go
package auctioninspect

import (
	"farental/core/data/api"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/widget/auctiondetails"
	"farental/widget/help"
	"farental/widget/iteminspect"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	details   *auctiondetails.Widget
	inspector *iteminspect.Widget

	help *help.Widget

	layout *layout.CenterLayout

	auction api.AuctionResponse
}

var _ orvyn.Screen = (*Screen)(nil)

func New(auction api.AuctionResponse) *Screen {
	s := new(Screen)

	s.auction = auction

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable("")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.details = auctiondetails.New()
	s.inspector = iteminspect.New()

	s.help = help.New()

	// The auction box holds a label and a right-flushed value; 0.40 of the 109
	// columns left after the layout margin and the gap gives it 43, which fits
	// the longest translated label. The compensator index is 1, so the rounding
	// remainder widens the item box rather than the auction one.
	elements := []layout.FixedRatioRenderable{
		layout.NewFixedRatioRenderable(0.40, s.details),
		layout.NewFixedRatioRenderable(0.60, s.inspector),
	}

	inspectLayout := layout.NewHBoxFixedRatioLayout(0, 1, 1, elements...)

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(10, t.Size(ftheme.LayoutWidthSizeID),
			orvyn.NewSize(10, 4),
			s.title,
			orvyn.VGap,
			inspectLayout,
			orvyn.VGap,
			s.help),
	)

	return s
}

func (s *Screen) OnEnter(any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextBackAndQuit)

	s.title.SetValue(fmt.Sprintf("%s x%d",
		s.auction.Item.Name, s.auction.Quantity))

	s.inspector.UpdateData(&s.auction.Item)
	s.details.UpdateData(&s.auction)

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
```

- [ ] **Step 2: Verify the module builds and every test passes**

Run: `go build ./... && go vet ./... && go test ./... && gofmt -l .`
Expected: all pass, `gofmt -l .` prints nothing.

Note the expected `go vet` silence in particular: it is what catches a leftover
unused import from the deleted `detailLines`.

- [ ] **Step 3: Check the dialog in the running client**

The TUI needs a real TTY and a live backend, so this step is the user's to run,
not the agent's. Ask the user to launch the client, open the auction house buy
screen, select an auction and press `i`.

Confirm, from the screenshot or their report:
1. two bordered boxes side by side, auction left, item right, tops and bottoms aligned;
2. the title reads `<item name> x<quantity>`;
3. auction values are flush to the right edge of the left box, labels dim, `Current bid` and `Direct buy` brighter than the other values;
4. an auction with no direct buy shows no `Direct buy` row and the box shrinks by one line;
5. an auction the player leads shows `(you)` after the bid;
6. `esc` closes the dialog and the buy screen's own help line comes back.

- [ ] **Step 4: Commit**

```bash
git add screen/dialog/auctioninspect/auctioninspect.go
git commit -m "feat: lay out the auction inspect dialog side by side"
```

---

## Done

Full verification: `go build ./... && go vet ./... && go test ./... && gofmt -l .`

Untouched on purpose, per the spec's out-of-scope list: `widget/iteminspect` and
therefore `screen/inventory` and `screen/shop`; scrolling for either box;
colouring `Ends in` when an auction is about to close; the item name appearing
both in the dialog title and as `iteminspect`'s own header.
