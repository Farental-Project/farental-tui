# Auction house buy screen

## Goal

Give the auction house a browse screen: filter the listings with the server's
filter vocabulary, page through the results on demand, inspect an item, bid on
it or buy it outright.

## Scope

- `widget/auctionfilter` — the filter panel (skeleton exists, empty today).
- `widget/auctionlistitem` — one auction row.
- `screen/auctionhouse/buy` — the screen (stub exists, renders nothing).
- `screen/dialog/statskillselection` — searchable stat-or-skill picker.
- `screen/dialog/auctionbid` — bid input.
- `screen/dialog/auctioninspect` — item and auction details.
- `internal/keybind` — help context for the screen.
- `internal/helper` — signed numeric input validation.
- `screen/auctionhouse/menu` — wire the Buy button, which returns nil today.

Out of scope: the manage screen (own listings and bids), auction search by
item name, saved filter presets.

## Current state

Server, `farental-cli` (branch `auctions-filters`, merged):

- `GET /auction/all` takes `page`, `itemID`, `kind`, `slot`, `weaponType`,
  `stat`, `skill`, `minStat`. Codes are resolved by the controller, so an
  unknown code is a readable 400 rather than an empty result. `stat` and
  `skill` together are rejected; `minStat` is ignored without one of them. The
  browsing character's own listings are excluded server-side. Page size is 50;
  the response carries `Total`, `Page`, `PageSize`.
- `GET /auction/filterOptions` returns `Kinds`, `EquipmentSlots`,
  `WeaponTypes`, `Stats`, `Skills`, each with a localized label. Skills with no
  primordial stat and inactive skills are withheld, matching what
  `/auction/all` accepts.
- `POST /auction/bid` takes `{ID, Bid}`; `POST /auction/buy` takes `{ID}`.

Client, `farental-tui`:

- `request.AuctionGetAll(page int, filter api.AuctionFilter)` sends only the
  filters that are set, and `request.AuctionGetFilterOptions()` fetches the
  vocabulary. `api.AuctionFilter`, `api.AuctionItemKind`,
  `api.AuctionFilterOptionsResponse` and `api.WeaponTypeResponse` mirror the
  server contract. Covered by `core/request/auction_test.go`.
- `screen.IDAuctionHouseBuy` is reserved and unused.
- `screen/auctionhouse/buy/buy.go` holds a `widgetlist.Widget[api.AuctionResponse]`
  and returns nil from every method.
- `widget/auctionfilter` holds an empty struct embedding `orvyn.BaseWidget` and
  `orvyn.BaseFocusable`.
- `widget/iteminspect` renders item details; used by the inventory and shop
  screens.
- `screen/generic/selectionlist` provides a searchable list screen;
  `screen/dialog/itemselection` shows how to embed it as a dialog.

## Structure

The screen follows the script editor: sibling focusable widgets owned by a
screen-level `FocusManager`, widgets reporting to the screen with typed
messages rather than reaching into it.

```
buy.Screen
├── title, characterInfo            (money is needed to judge a bid)
├── HBoxFixedRatioLayout 0.25 / 0.75
│   ├── auctionfilter.Widget        focusable
│   └── widgetlist[AuctionResponse] focusable
├── statusMessage
└── help
```

`focusManager.SetWidgets([]orvyn.Focusable{auctionFilter, list})`, Tab and
Shift+Tab between the two. The filter runs its own `FocusManager` over its five
controls, as `auctionstartform` does.

### auctionfilter.Widget

| Control | Widget | Source |
| --- | --- | --- |
| Kind | `multivalueselector[Option]` | `options.Kinds` |
| Equipment slot | `multivalueselector[Option]` | `options.EquipmentSlots` |
| Weapon type | `multivalueselector[Option]` | `options.WeaponTypes` |
| Stat or skill | `button` → picker dialog | `options.Stats`, `options.Skills` |
| Minimum stat | `textinput` | typed |

`Option{Code, Label}` implements `multivalueselector.Value` through
`RenderValue()`. Every selector starts with an empty-code option rendered as
`—`, meaning "no filter".

Public surface:

- `SetOptions(*api.AuctionFilterOptionsResponse)` — fills the selectors.
- `GetFilter() api.AuctionFilter` — reads the controls and delegates to a pure
  `buildFilter`.
- `Reset()` — back to the all-empty filter.
- `AppliedMsg` + `AppliedCmd` — emitted on Enter, in the style of
  `scriptrulelist.FocusInspectorMsg`.

Stat and skill share one control, so the server's stat-and-skill rejection is
unreachable from this client. `minStat` sets `HasMinStat` only when the text
parses and a stat or skill is picked; the request layer omits it in every other
case, so the two agree.

The minimum accepts negative values — penalties are legal content — which
`helper.NumericalValidate` does not allow. A `helper.SignedNumericalValidate`
(`^-?\d*$`) is added beside it.

### statskillselection dialog

Embeds `selectionlist.Screen[Entry]` like `itemselection` does, listing stats
then skills with a header for each. Returns one
`Entry{Kind: stat|skill, Code, Label}` on submit, nothing on cancel.

### auctionlistitem

One row per `api.AuctionResponse`:

```
> Steel Helm      x1  bid 120₲ (you) buy 300₲  4h02  Bob
  Iron Helm       x2  bid  40₲       buy   —   12h   Alice
```

`(you)` marks the character as current bidder. `—` replaces the buy price when
`DirectBuyPrice` is 0. The ends-in column formats `EndTimestamp` as `4h02`,
`12h`, `1d 3h`, or an expired marker.

### auctioninspect dialog

`widget/iteminspect` for the item, plus seller, current bidder, duration and
end time.

### auctionbid dialog

Numeric input prefilled with `CurrentBid + 1`, showing the item and the
character's grynars. Submits `api.AuctionBidBody{ID, Bid}`.

## State and flows

```go
type Screen struct {
	// widgets…

	options *api.AuctionFilterOptionsResponse
	filter  api.AuctionFilter
	page    int
	total   int64
}
```

`filter` is the *applied* query; the widget owns the one being edited. Load-more
must page the query that produced the current rows, not whatever the selectors
have drifted to since.

- `OnEnter` — switch help context, refresh character info, fetch the filter
  options into the widget, `Reset()` it, then run the apply flow with the empty
  filter.
- Apply (`AppliedMsg`) — `filter = auctionFilter.GetFilter()`, `page = 1`,
  fetch, `list.SetItems(resp.Auctions)`, cursor to the first row, `total` from
  the response.
- Load more (`m`) — `page++`, fetch with the stored `filter`, append to
  `list.GetItems()`. A no-op with a status line once
  `int64(len(items)) >= total`.
- Inspect (`i`) — open `auctioninspect` for the selected row.
- Bid (Enter on the list) — open `auctionbid`; on submit `POST /auction/bid`.
- Buy (`b`) — yes/no confirm showing the price, then `POST /auction/buy`.
  Hidden when `DirectBuyPrice` is 0.
- After a successful bid or buy — re-run the apply flow from page 1. A sold
  listing is gone and every later page shifts under it, so refetching the
  accumulated pages would show duplicates or holes. Accumulation resets and the
  status message reports the outcome.
- Reset (`r`) — clear every filter control and re-apply at once, returning the
  player to the whole house in one keypress. Handled at screen level so it
  works from either panel: a filter is most often reset while editing it.
- Esc — back to the auction house menu.

Keys, list focused: Enter bid, `b` buy, `i` inspect, `m` load more. Either
panel: `r` resets the filter. Filter focused: ←/→ cycle, Space opens the
picker, Enter applies.

## Errors

- All calls go through `helper.Fetch` / `helper.SendRequest`; failures land in
  `statusMessage.SetError`. Filter rejections already arrive as joined
  localized sentences, so no per-field plumbing is needed.
- A failed `filterOptions` fetch shows the error and leaves the selectors
  empty; the list still browses unfiltered. It is refetched on every `OnEnter`,
  since the labels follow the player's language.
- No result is a status line, not an error — an empty house and a bad filter
  must not look alike.
- 401 handling is untouched: a messageless 401 expires the session, a business
  401 ("no auction house at this location") carries its message.

## Testing

Minimal, and only where the logic is pure — widget constructors need a theme,
so they are not unit-tested, consistent with the rest of the repo:

- `buildFilter` — fields set and unset, minimum with no stat, negative minimum.
- The ends-in formatter — hours, days, expired.

`core/request/auction_test.go` already covers query-param omission. Everything
else is verified with `go build ./...`, `go vet ./...`, `gofmt -l` and
`go test ./...`. The screen itself needs a TTY and a running backend, so it is
not exercised automatically; that gap is stated rather than papered over.

## Conventions

Comments explain why, not what, and stay short. No comment restates a line of
code below it.
