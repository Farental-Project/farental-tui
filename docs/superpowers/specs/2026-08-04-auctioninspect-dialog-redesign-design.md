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
- the dialog is one tall column, wasting the 120-column width it is given.

## Target layout

```
                              Iron Sword x3

┌──────────────────────────────┐ ┌───────────────────────────────────────────┐
│ Auction                      │ │ Iron Sword                                │
│ ───────                      │ │ ──────────                                │
│ Current bid          120Ǥ    │ │ 10 per stack                              │
│ Direct buy           500Ǥ    │ │                                           │
│ Current bidder        you    │ │ A sturdy blade forged in the northern     │
│ Seller           Jane Doe    │ │ forges.                                   │
│ Ends in            2h 30m    │ │                                           │
│                              │ │ Stats                                     │
│                              │ │ ─────                                     │
│                              │ │ • Attack : 12                             │
└──────────────────────────────┘ └───────────────────────────────────────────┘

                          esc back · q quit
```

Two side-by-side boxes: auction left, item right. Both stretch to the same
height, so the outlines align regardless of how much content each holds.

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
- `Render() string` — has `GetContentSize()` available, which is what makes the
  right-flush column possible at all. Each row splits the content width into a
  label column and a value column, renders the label left-aligned and the value
  right-aligned, and `ansi.Truncate`s both to their column budget first.
  Truncation is required, not cosmetic: `lipgloss` `Width(n)` wraps rather than
  truncates, and a wrapped row would push the box past its allotted height (the
  same reasoning as `widget/auctionlistitem/auctionlistitem.go:99`). The final
  render applies `Width` **and** `Height` from the content size, as
  `iteminspect.Render` does, so the box fills its half of the row and the two
  outlines line up.
- `GetMinSize()` and `GetPreferredSize()` both return
  `orvyn.NewSize(1, headerHeight + len(rows) + GetStyle().GetVerticalFrameSize())`.
  Width stays at 1 so the widget never drives the layout width, matching
  `iteminspect.GetMinSize`.

Rows, in order:

| Label            | Value                                        | Shown when          |
|------------------|----------------------------------------------|---------------------|
| `Current bid`    | `<n>Ǥ`                                        | always              |
| `Direct buy`     | `<n>Ǥ`                                        | `DirectBuyPrice > 0`|
| `Current bidder` | `you` when the player leads, `nobody` when empty, otherwise the bidder name | always |
| `Seller`         | `SellerName`                                  | always              |
| `Ends in`        | `auctionlistitem.EndsIn(EndTimestamp, now)`   | always              |

The own-bid marker lives on the bidder row rather than as a ` (you)` suffix on
the price: the question it answers is *who* holds the bid, so it belongs in the
column that names the bidder, and the price column stays purely numeric.

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

### 4. Dialog wiring

`screen/dialog/auctioninspect` keeps its `CenterLayout` +
`DefinedWidthVerticalLayout`, but the two inspectors move into a horizontal row:

```go
elements := []layout.FixedRatioRenderable{
	layout.NewFixedRatioRenderable(0.40, s.details),
	layout.NewFixedRatioRenderable(0.60, s.inspector),
}

inspectLayout := layout.NewHBoxFixedRatioLayout(0, 1, 1, elements...)
```

Vertical children become `title`, `orvyn.VGap`, `inspectLayout`, `orvyn.VGap`,
`help` — the same shape `screen/dialog/fightinspector/fightinspector.go:75`
uses.

Ratio rationale: the layout is 120 wide (`ftheme.LayoutWidthSizeID`,
`internal/theme/farentaldark.go:128`) less a 10-wide margin and a 1-column gap,
leaving 109 to split. 0.40/0.60 gives the auction box 43 columns (41 inner) and
the item box 65 plus the 1-column rounding remainder, so 66 (64 inner). The
compensator index is 1, so the remainder goes to the item box.

41 inner columns is sized off the longest translated label rather than the
English one: `Current bidder` is 14 characters, but the French
`Enchérisseur actuel` is 19 and the German `Aktueller Bieter` is 16. At a 0.55
label / 0.45 value split that leaves 22 columns for the label and 19 for the
value, which fits every label and a full character name untruncated.

`detailLines()` is deleted; its content is now the widget's concern.

- `OnEnter` — unchanged except that it sets the `xN` title and calls
  `details.UpdateData` instead of `details.SetValue`. The keybind context stays
  `keybind.ContextBackAndQuit`.
- `Update` — unchanged. `Esc` closes the dialog; nothing else is interactive.

`iteminspect` is not modified at all, so `screen/inventory` and `screen/shop`
are untouched.

### 5. Translations

One new string, `"Auction"`, added to the en / fr / de catalogues. Every other
label (`Current bid`, `Direct buy`, `Current bidder`, `Seller`, `Ends in`,
`nobody`, `you`) already exists.

## Testing

A table test for `auctiondetails` row building, next to the existing
`widget/auctionfilter/filter_test.go` and
`widget/auctionlistitem/endsin_test.go`:

- `DirectBuyPrice == 0` omits the Direct buy row; `> 0` includes it;
- an empty `CurrentBidderName` renders `nobody`;
- the row count that `GetMinSize` reports matches the rows actually built, in
  both the with- and without-direct-buy cases — this is the invariant that keeps
  the box from overflowing its allocated height.

Row construction is split from `Render` into a pure function so the test needs
no terminal.

Then `go build ./... && go vet ./... && go test ./... && gofmt -l .`.

### 6. Both boxes stretch to fill the dialog height

Left to themselves the two boxes are only as tall as their content and sit
centred, wasting vertical space. `resizeFlexibleElements`
(`orvyn/layout/flexheight.go:37`) treats an element as fixed height when its
preferred height equals its minimum, and both boxes report exactly that:
`auctiondetails` by construction, `iteminspect` because `SimpleRenderable`
returns the same measurement for min and preferred (`orvyn/simple.go:62-80`).
`HBoxFixedRatio` takes the max over its children, so the whole row reads as
fixed and the vertical layout hands it only its content height.

So `auctiondetails.GetPreferredSize` honours an explicitly-set preferred size,
falling back to the row-count height when none was set — the same override
idiom `SimpleRenderable` uses. The dialog then calls
`SetPreferredSize(orvyn.NewSize(1, stretchedHeight))` on it, which makes the box
row the only flexible child of the vertical layout, so it receives all the
leftover height. `HBoxFixedRatio.Render` already resizes both children to the
full row height (`orvyn/hfixedratio.go:72`), so the two boxes grow together and
their outlines stay aligned.

`stretchedHeight` is 200. Its magnitude does not matter — a sole flexible
element receives the entire surplus whatever its weight — so it only has to
exceed the box's own row count, and the layout clamps it to the height the
terminal actually has.

### 7. Both boxes clip to the height they are allocated

Stretching the row exposed a contract both boxes were quietly breaking:
`lipgloss` `Height(n)` pads but never truncates, and `VBoxLayout.Render`
(`orvyn/layout/vbox.go:70-73`) renders its children at their natural size
without consulting what it was allocated. So a box handed less height than its
content needed simply drew past its own border.

`HBoxFixedRatio` gives both children the *same* height, so alignment requires
both to honour it. Each now clips its inner content before the bordered style
wraps it:

```go
	if contentSize.Height > 0 {
		content = lipgloss.NewStyle().
			MaxHeight(contentSize.Height).
			Render(content)
	}
```

`MaxHeight` truncates where `Height` pads. The `> 0` guard is required:
`MaxHeight(0)` is a pass-through in lipgloss (`style.go:461`), not a
clip-to-empty, and a zero height is a legitimate transient during layout. The
clip goes on the inner content, never the outer bordered style — clipping the
outer style would cut the bottom border and leave the box unclosed.

This is why `iteminspect` had to change despite being shared: `screen/inventory`
and `screen/shop` both size it through an `HBoxFixedRatio` at ratio 0.40 with a
real positive height, so they now clip long descriptions where they previously
overflowed their panels.

Verified headless at 120 columns: at heights 50, 40, 29, 24, 20, 16 and 13 the
dialog renders exactly the requested number of lines with both outlines aligned.

## Known limitations

- Neither box scrolls. An item with a long description or many stats has its
  tail cut off rather than being reachable. Same for `screen/inventory` and
  `screen/shop`, which share `iteminspect`.
- Below about 13 rows the dialog's own minimum no longer fits and it overflows
  the terminal — at 120×10 it renders 26 lines. The title and help line are
  fixed height, so only the box row can absorb a shortfall, and past a point
  there is nothing left to give. This is below the size at which the rest of the
  app is usable.
- The box header (`Auction`) is not truncated before `Width()` is applied, so at
  content widths of 1-6 columns it wraps and the box draws taller than
  `GetMinSize` reports. Reachable only below about 30 terminal columns. Reviewed
  and accepted deliberately.
- `"Quantity"` is now an orphaned translation key in all three catalogues: the
  deleted `detailLines()` was its only caller. Left in place rather than removed
  as part of this change.
- `orvyn`'s `resizeFlexibleElements` gives a sole flexible element `max(left, 0)`
  with no minimum floor, contradicting its own doc comment. Commit `1b7e017` in
  the orvyn repository adds the floor, but it is a no-op in the case that
  motivated it: when the minimums do not fit, `guaranteed` is rebound to return
  0 (`flexheight.go:117`), so the floor collapses back to zero. The fix is kept
  because it closes a real latent bug for multi-element flexible rows.

## Out of scope

- Colouring `Ends in` when an auction is about to close. It needs a threshold
  decision and a re-render tick.
- Removing the duplicated item name between the dialog title and `srName`.
