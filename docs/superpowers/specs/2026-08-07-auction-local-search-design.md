# Local search in the auction house filter

## Goal

Add a free-text field to the auction house filter panel that narrows the
auction list by fuzzy match on the item name. The narrowing is local: it acts
on the rows already fetched and never reaches the server.

## Behaviour

The field sits at the top of the filter panel, above `Show`, and is the panel's
first tab stop.

Enter keeps its single meaning across the whole panel: refetch with the
server-side criteria, then narrow the result with the search text. There is no
focus-dependent Enter.

`Ctrl+R` (reset) clears the search text along with every other control. `r`
(reload) and the reloads that follow a bid or a buy re-apply the *committed*
search, not an unconfirmed edit — the same applied-versus-being-edited rule the
panel's other controls already follow.

The status line reports `shown/loaded/total` whenever a search is applied
(`7/40/312 auction(s)`), including when the text happens to match every loaded
row, and falls back to the existing `loaded/total` (`40/312 auction(s)`) when
the applied search is empty. Keeping `loaded` visible is what makes `m` (load
more) still legible.

When the search matches nothing, the list is empty and the existing
"No auction matches these filters" message applies unchanged.

## The single-letter keybind collision

`buy.go` handles `m`, `r`, `b` and `i` at screen level *before* delegating to
the focus manager. Today that is safe, and the code says so:

> Inspecting is read-only, so it stays available while the filter panel holds
> focus. Nothing there can swallow the key: the only text input is
> numeric-validated.

A free-text field makes that false — typing `sword` would fire reload, buy
confirm and load more. So:

- `auctionfilter` exposes `SearchFocused() bool`.
- `buy.go` guards the four cases on it and hides the four bindings from the
  help line through the existing `updateFocusKeybinds`.
- The stale comment is rewritten.

The field is left with Up/Down, which the panel already binds to move between
controls (Tab still leaves the panel entirely). Moving off the field restores
the letter keys. Esc keeps its one meaning — leave the screen.

The rejected alternative was to give the field a real orvyn input mode. The
panel is a single `Focusable` in the screen's focus manager, so its
`IsInputting()` would make `FocusManager.Update` route every message into the
panel and claim Esc as the exit-input key, fighting the screen's Esc-to-leave.
Guarding on focus is what the codebase already does elsewhere
(`scripteditor.go:221`).

## orvyn changes (`widget/widgetlist`)

The list already owns fuzzy matching (`FuzzyFilter`, `filteredListItems` and
the index mapping behind `GetSelectedItem`). It is reused rather than
reimplemented; only the door is missing.

`SetFilterable(false)` already deactivates the built-in text input and drops it
from the height maths, so the panel can own the visible field with no further
change.

| Change | Reason |
| --- | --- |
| `SetFilter(s string)` | Public entry point. Sets `tiFilter`'s value, then runs `filter(s)`, so the value stays in sync with the applied filter and the re-apply paths in `SetItem`/`AppendItem` keep working. |
| `VisibleLength() int` | `Length()` returns the unfiltered count. The `Length() > 0` guards and the count line would both lie while a filter is applied. |
| `filter("")` returns after `clearFilter()` | Missing `return`: it clears, then immediately re-filters with `""` and sets `FilterApplied`. |
| `SetItems` re-applies an applied filter | `SetItem` and `AppendItem` do; `SetItems` does not, leaving `filteredListItems` holding indices into the previous slice. Same class as the stale-index panics already fixed in the selection lists. |
| Esc-clear gated on `filterable` | `Update` clears the filter on Esc even when `filterable` is false, which would desync the panel's field from the list. Unreachable before this change, since `filter` was private and only the `/` key reached it. |

## farental changes

### `widget/auctionfilter`

- `lblSearch` and `tiSearch` first in the layout and first in the focus
  manager. No `Validate` — the field takes any text.
- `GetSearch() string` and `SearchFocused() bool`.
- `Reset()` clears the field.
- The text is deliberately absent from `GetFilter()`: it never reaches the
  server.
- Label and placeholder localized through `lokyn.L`, refreshed in `Init` like
  the panel's other labels.

### `screen/auctionhouse/buy`

- A `search` field beside `filter`, holding the applied search for the same
  reason `filter` holds the applied query.
- `reload(filter, search, mayReport)` calls `SetFilter(search)` after
  `SetItems`.
- `applyFilter` passes `auctionFilter.GetSearch()`.
- The `Length() > 0` guards before inspect, bid and buy become
  `VisibleLength() > 0`.
- `reportCount` gains the three-part form.
- The four single-letter cases gain the `SearchFocused()` guard and
  `updateFocusKeybinds` gains their visibility.

### `translations/{en,de,fr}.json`

The label and the placeholder.

## Testing

orvyn `widget/widgetlist`:

- `SetFilter` narrows the list and `GetSelectedItem` maps through the filtered
  indices.
- `SetFilter("")` returns the widget to `Unfiltered`.
- `SetItems` under an applied filter re-filters instead of stranding indices
  into the old slice — including the shrinking case that panicked before.
- `VisibleLength` tracks the filtered count while `Length` keeps reporting the
  full one.

farental `widget/auctionfilter`:

- `GetSearch` reads the field back.
- `GetFilter` ignores the search text.
- `Reset` clears the field.
