# Auction inspect dialog redesign

Date: 2026-08-04
Branch: `auction-buy-screen`

## Problem

`screen/dialog/auctioninspect` renders the item inspector and the auction facts
as two unrelated blocks. The auction facts are a single `SimpleRenderable` fed a
`"label: value\n"` string built by `detailLines()`, so:

- values do not line up in a column, they float wherever the label ends;
- there is no visual boundary between the item and the auction, even though
  `iteminspect` already draws a border (`orvyn.NewBaseWidget()` sets
  `BlurredWidgetStyleID`, `orvyn/widget.go:43`) — the auction text sits outside
  it and reads as an afterthought;
- a long item description, a long stat list, or a small terminal clips the
  inspector with no way to reveal the hidden lines.

## Target layout

```
                    Iron Sword x3

┌──────────────────────────────────────────────┐
│ Iron Sword                                   │
│ ──────────                                   │
│ 10 per stack                               ↑ │
│                                              │
│ A sturdy blade forged in the northern forges.│
│                                              │
│ Stats                                        │
│ ─────                                        │
│ • Attack : 12                              ↓ │
└──────────────────────────────────────────────┘
┌──────────────────────────────────────────────┐
│ Auction                                      │
│ ───────                                      │
│ Current bid                       120Ǥ (you) │
│ Direct buy                             500Ǥ  │
│ Current bidder                      John Doe │
│ Seller                              Jane Doe │
│ Ends in                               2h 30m │
└──────────────────────────────────────────────┘

              ↑↓ scroll · esc back · q quit
```

Two stacked full-width boxes. The item box flexes to fill the dialog and
scrolls; the auction box is fixed to exactly the height of its rows.

## Components

### 1. `widget/auctiondetails` (new)

A peer of `widget/iteminspect`. The package is named `auctiondetails`, not
`auctioninspect`, because the dialog package is already called `auctioninspect`
and importing a same-named package into it would need an alias at every call
site.

```go
type row struct {
	label string
	value string
	style lipgloss.Style
}

type Widget struct {
	orvyn.BaseWidget

	header *orvyn.SimpleRenderable
	rows   []row
}
```

- `New() *Widget` — the border needs no work: `orvyn.NewBaseWidget()` already
  applies `BlurredWidgetStyleID`, so the box matches `iteminspect` exactly.
  `header` holds `lokyn.L("Auction")` styled with
  `ftheme.DimUnderlinedTextStyleID`, the same section-title treatment
  `iteminspect` uses for `Stats` / `Equip conditions` / `Effects`.
- `UpdateData(auction *api.AuctionResponse)` — rebuilds `rows`.
- `Render() string` — has `GetContentSize().Width` available, which is what
  makes the right-flush column possible at all. Each row splits the content
  width into a label column and a value column, renders the label left-aligned
  and the value right-aligned, and `ansi.Truncate`s both to their column budget
  first. Truncation is required, not cosmetic: `lipgloss` `Width(n)` wraps
  rather than truncates, and a wrapped row would push the box past the height it
  reported (the same reasoning as `widget/auctionlistitem/auctionlistitem.go:99`).
- `GetMinSize()` and `GetPreferredSize()` both return
  `orvyn.NewSize(1, headerHeight + len(rows) + GetStyle().GetVerticalFrameSize())`.
  Equal min and preferred makes `isFixedHeight` true
  (`orvyn/layout/flexheight.go:37`), so the layout hands the box exactly its own
  height and leaves the rest to the item box. Width stays at 1 so the widget
  never drives the layout width, matching `iteminspect.GetMinSize`.

Rows, in order:

| Label            | Value                                        | Shown when          |
|------------------|----------------------------------------------|---------------------|
| `Current bid`    | `<n>Ǥ`, plus ` (you)` when the player leads   | always              |
| `Direct buy`     | `<n>Ǥ`                                        | `DirectBuyPrice > 0`|
| `Current bidder` | bidder name, or `nobody` when empty           | always              |
| `Seller`         | `SellerName`                                  | always              |
| `Ends in`        | `auctionlistitem.EndsIn(EndTimestamp, now)`   | always              |

Styling: labels use `theme.DimTextStyleID`. `Current bid` and `Direct buy`
values use `theme.HighlightTextStyleID` so money reads first; all other values
use the normal text style.

`Quantity` is deliberately absent — see section 3.

### 2. `auctionlistitem.IsOwnBid` (extracted)

`widget/auctionlistitem/auctionlistitem.go:59` decides whether the player holds
the current bid by comparing `CurrentBidderName` against
`context.CharacterInfo` (the server sends readable names, not IDs). That
comparison moves into an exported helper in the same package, next to the
already-exported `EndsIn`:

```go
func IsOwnBid(bidderName string) bool
```

`auctionlistitem.UpdateData` and `auctiondetails` both call it. One source of
truth, so a future switch to IDs is a single edit.

### 3. Quantity moves into the dialog title

The title becomes `"<item name> x<quantity>"`, matching
`screen/dialog/auctionbid/auctionbid.go:84` and
`widget/auctionlistitem/auctionlistitem.go:76`. The `Quantity` row leaves the
auction box.

The item name still appears twice — once in the dialog title, once as
`iteminspect`'s `srName` header inside the box. That is pre-existing and stays;
the `xN` suffix at least makes the two lines carry different information.

### 4. `widget/iteminspect`: opt-in scrolling

Scrolling is added behind a flag so `screen/inventory` and `screen/shop`, which
share this widget, keep today's behaviour byte-for-byte. Their identical
clipping problem is real but out of scope here.

New state: `viewport viewport.Model`, `scrollable bool`.

- `SetScrollable(bool)` — off by default.
- `Resize` — when scrollable, sizes the viewport to the content size.
- `Render` — when not scrollable, the current code path is untouched. When
  scrollable, the layout render goes through `helper.SetScrollableContent` and
  the result through `helper.OverlayScrollArrows` (`internal/helper/scroll.go`),
  the same pair `simplelogviewer` uses, so the up/down arrow affordance is
  identical to every other scrollable box in the app.
- `Update` — when scrollable, `up` / `down` scroll the viewport by one line.
  When not scrollable it keeps returning nil.
- `UpdateData` — when scrollable, `GotoTop()` after refreshing, so a new item
  never opens mid-scroll.
- `GetMinSize` — when scrollable, the height is
  `min(3 + frame, currentBehaviour)` instead of the full inner-layout height:
  three content lines is enough to be readable and to show both scroll arrows,
  and the `min` keeps a short item from reserving more than it needs. Min then
  differs from preferred, which is what makes `isFixedHeight` false and lets the
  box absorb the dialog's leftover height. When not scrollable it returns
  exactly what it returns today.
- `GetPreferredSize` — unchanged in both modes.

### 5. Dialog wiring

`screen/dialog/auctioninspect`:

```
title       SimpleRenderable, TitleStyleID   fixed
inspector   iteminspect, SetScrollable(true) flexible
details     auctiondetails.Widget            fixed
help
```

`detailLines()` is deleted; its content is now the widget's concern.

- `OnEnter` — `bubblehelp.SwitchContext(keybind.ContextAuctionInspect)`, set the
  title, `inspector.UpdateData`, `details.UpdateData`.
- `Update` — `Esc` closes the dialog; every other message forwards to
  `inspector` so up/down scroll.

### 6. Keybind context

New `ContextAuctionInspect` registered in `internal/keybind/context.go` with
`Up`, `Down`, `Esc`, `Quit`.

`screen/auctionhouse/buy/buy.go:189` already calls
`bubblehelp.SwitchToPreviousContext()` on `dialogIDInspect` exit, so restoring
the buy-screen context needs no change.

### 7. Translations

One new string, `"Auction"`, added to the en / fr / de catalogues. Every other
label (`Current bid`, `Direct buy`, `Current bidder`, `Seller`, `Ends in`,
`nobody`, `you`) already exists.

## Testing

A table test for `auctiondetails` row building, next to the existing
`widget/auctionfilter/filter_test.go` and
`widget/auctionlistitem/endsin_test.go`:

- `DirectBuyPrice == 0` omits the Direct buy row; `> 0` includes it;
- an empty `CurrentBidderName` renders `nobody`;
- the row count that `GetMinSize` reports matches the rows actually rendered,
  in both the with- and without-direct-buy cases — this is the invariant that
  keeps the box from overflowing its allocated height.

Row construction is split from `Render` into a pure function so the test needs
no terminal.

Then `go build ./... && go vet ./... && go test ./... && gofmt -l .`.

## Out of scope

- Scrolling for `iteminspect` in `screen/inventory` and `screen/shop`. The
  widget will support it; wiring the keys and focus there is a separate change.
- Colouring `Ends in` when an auction is about to close. It needs a threshold
  decision and a re-render tick.
- Removing the duplicated item name between the dialog title and `srName`.
