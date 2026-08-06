# Auction house manage screen and outbid tracking

## Goal

Two related gaps close together.

A player who is outbid learns about it by mail, then has to find the auction
again by hand in a house of thousands. There is no way to ask "which auctions
did I bid on and lose the lead in". This adds one.

Separately, the auction house menu offers Manage and the button does nothing.
The player cannot see their own listings, cancel one, or check the bids they
are currently winning.

## Scope

Server, `farental-cli`:

- `model/data/auction.go` — the participation record.
- `config/database.go` — register it for AutoMigrate.
- `repository/auction.go` — write it, and filter on it.
- `serverapi/controller/auctiontrade.go` — record a bid as it is placed.
- `serverapi/controller/auction.go` — the `outbid` browse parameter.
- `serverapi/system/kic/keepitclean.go` — purge settled auctions.
- `migrations/structure/s1.5.0__auctions.sql` — the new table.

Client, `farental-tui`:

- `core/data/api/auction.go`, `core/request/auction.go` — the `outbid` param.
- `widget/auctionfilter` — the scope control.
- `screen/auctionhouse/manage` — the screen (directory exists, empty).
- `screen/auctionhouse/menu` — wire the Manage button, which returns nil today.
- `internal/keybind` — help context for the screen.
- `main.go` — register the screen.

Out of scope: bid history with amounts, a notification of being outbid (mail
already does that), acting on a winning bid from the manage screen.

## Deferred

The manage screen's bids pane is look-only. A player who is outbid while
looking at it sees the row vanish on the next refresh, with nowhere on that
screen to respond — they have to cross to the buy screen and switch the filter.

Merging the two panes' concerns, or giving the manage screen a toggle between
winning and outbid, was considered and deliberately postponed: this piece of
work is large enough already. Revisit once the three auction screens are all
in place and the real navigation cost is observable rather than guessed.

## Current state

### The server cannot answer the question

`data.Auction` carries a single `CurrentBidderID`. The moment someone outbids
you, the row stops referencing you at all, and no other table records that you
were ever there:

- `/auction/bids` is `current_bidder_id = ? AND status = Active AND
  end_timestamp > now` (`repository/auction.go:219`). By construction it returns
  only the auctions you are *winning*.
- `auction.RefundBidder` mails the money back with the item name and quantity,
  but no auction ID. A mailbox is not a queryable index.
- There is no `GET /auction/:id`, so a client cannot even re-resolve auction IDs
  it remembered locally.

So the feature needs a server-side record of participation. A client-local
watchlist was rejected: it dies on reinstall, does not follow the player to
another client, and has no endpoint to resolve against.

### kic does not clean auctions

`keepitclean.go` sweeps fights, orphaned pending fight results, mail, finished
tasks, event log entries, stat modifiers and chat messages. Auctions are absent,
so every Sold, Expired and Cancelled row lives forever. Pre-existing, and the
new table would inherit it.

### The client

- `request.AuctionGetOwn()` and `request.AuctionGetBids()` exist and are called
  from nowhere. The manage screen is their first consumer.
- `screen.IDAuctionHouseManage` is reserved and unused;
  `screen/auctionhouse/manage/` is an empty directory.
- `menu.go` handles `SKey` and `BKey`; there is no `MKey` case, and
  `btManageOnClicked` returns nil.
- `widget/auctionlistitem`, `screen/dialog/auctioninspect` and
  `screen/dialog/popup` are all reusable as they stand.

## Part 1 — the participation record

### Model

```go
// AuctionBid records that a character has bid on an auction. It carries no
// amount: an outbid player was already refunded by mail, so their old bid is
// dead money, and the current bid — the number that decides whether to bid
// again — is on the auction itself.
//
// Both foreign keys cascade, unlike Auction's own. Auction deliberately blocks
// deleting a character with goods or a standing bid in play; a participation
// row holds neither, so it must never block deleting the auction it belongs to
// or the character that wrote it.
type AuctionBid struct {
	AuctionID uint    `gorm:"primaryKey"`
	Auction   Auction `gorm:"constraint:OnDelete:CASCADE"`

	CharacterID uint      `gorm:"primaryKey;index"`
	Character   Character `gorm:"constraint:OnDelete:CASCADE"`
}
```

The composite primary key *is* the uniqueness constraint. No surrogate id, and
re-bidding on the same auction needs no dedupe branch.

`AutoMigrate` takes an explicit model list rather than discovering them, so
`data.AuctionBid{}` is registered in `config/database.go` beside
`data.Auction{}`. Without it, development — the only environment AutoMigrate
runs in — never creates the table, and only production would have it.

Why no `amount` column: it would duplicate what `auctions.current_bid` already
holds for the leader, and for everyone else it records money that has already
been refunded. The auction row stays the authority on current state; this table
only answers "was this character ever here".

The `auctions` columns stay where they are. `current_bid` in particular cannot
be derived from this table: `plan.go:66` sets it to the seller's asking price at
listing, before any bid exists, and `auctiontrade.go:66-70` reads it that way.
Deriving it would also cost a lateral join per row on the hottest query in the
system, and move the bid invariant off the row that `LockAndReload` locks.

### Write

In `AuctionMakeBid`, immediately after the auction row is saved and inside the
same transaction:

```go
if err = ctx.Auctions.RecordBid(auctionRecord.ID, ctx.Character.ID); err != nil {
	slog.Error("auction make bid while recording participation", "err", err,
		"auctionID", auctionRecord.ID)
	ctx.DB.Rollback()
	return c.SendStatus(fiber.StatusInternalServerError)
}
```

`RecordBid` inserts through `Clauses(clause.OnConflict{DoNothing: true})`, so a
second bid on the same auction is a no-op.

Placement follows the rule the event log entry already documents: inside the
transaction, before `Commit()`. A rollback must take the participation row with
it, or the outbid list will offer an auction the player never actually bid on.

Failure rolls the whole bid back rather than committing an untracked one — every
other write in that handler behaves the same way, and the alternative is a bid
that silently drops out of the feature that exists to track it.

### Read: a filter, not an endpoint

`repository.AuctionFilter` gains one field:

```go
// OutbidFor narrows to auctions this character has bid on and no longer leads.
OutbidFor uint
```

applied in `FindAllWithPreloads`'s `scope` closure — it is a condition on the
auction, not on the item, so it does not belong in `itemScope`:

```go
if f.OutbidFor != 0 {
	bids := r.DB.Model(&data.AuctionBid{}).
		Select("auction_id").
		Where("character_id = ?", f.OutbidFor)

	// current_bidder_id is never NULL here: a bid row exists only because
	// this character led at some point, and nothing clears the leader while
	// the auction is Active — a later bidder replaces them, and every other
	// outcome sets a terminal status.
	db = db.Where("id IN (?)", bids).
		Where("current_bidder_id <> ?", f.OutbidFor)
}
```

Living in `scope` means `Count` and the page query receive it identically, so
`Total` stays truthful and paging keeps working.

A dedicated `/auction/outbid` endpoint was rejected on that basis: it would have
had to re-implement paging, `Total` and the existing filters, and the client
would have needed a second list mode. Being outbid is a property of an auction,
which makes it a filter.

### Controller

`parseAuctionFilter` reads the parameter:

```go
if raw := c.Query("outbid", ""); raw != "" {
	value, err := strconv.ParseBool(raw)

	if err != nil {
		invalid("outbid", raw)
	} else if value {
		filter.OutbidFor = ctx.Character.ID
	}
}
```

`c.QueryBool` is deliberately avoided, for the reason already written above
`minStat`: it maps an unparsable value to its default, which here would answer a
malformed request with the entire unfiltered house.

The swagger block on `AuctionGetAll` gains the parameter and its description.

`/auction/bids` and `/auction/own` are untouched. Part 3 is their first consumer.

### kic

A new step, following the shape of the others:

```go
// AUCTIONS
// Delete old settled auctions (30 days). Their auction_bids rows go with
// them through ON DELETE CASCADE.
//
// A cancelled auction's end_timestamp is still in the future, so it outlives
// the window by up to its remaining duration. Harmless at 96h worst case, and
// the alternative is a settled_timestamp column nothing else needs.
limitDate = time.Now().AddDate(0, 0, -30)

result = db.Where("status <> ? AND end_timestamp < ?",
	data.AuctionStatusActive, limitDate).Find(&oldAuctions)
```

then the same `Find`-check-`Unscoped().Delete()`-`logStep` sequence as its
neighbours, with `oldAuctions` declared alongside the other slices at the top of
the loop.

Nothing references an auction row after settlement — mail carries its own copies
of the money and the items — so the delete is safe.

### Migration

The table goes into `s1.5.0__auctions.sql` rather than a new file: `s1.5.0` is
not in production yet, the guard still requires structure `1.4.0`, and prod will
receive both tables in one pass. Development databases already at `1.5.0` pick
the table up from AutoMigrate.

The file's own instruction applies — generate the DDL from `SHOW CREATE TABLE`
against an AutoMigrated database rather than hand-writing it, so a database
built from the models and one built from the migration stay identical. Its
header description gains `auction_bids`.

Expected shape:

```sql
CREATE TABLE IF NOT EXISTS `auction_bids` (
  `auction_id`   bigint(20) unsigned NOT NULL,
  `character_id` bigint(20) unsigned NOT NULL,
  PRIMARY KEY (`auction_id`,`character_id`),
  KEY `idx_auction_bids_character_id` (`character_id`),
  CONSTRAINT `fk_auction_bids_auction` FOREIGN KEY (`auction_id`)
    REFERENCES `auctions` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_auction_bids_character` FOREIGN KEY (`character_id`)
    REFERENCES `characters` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;
```

## Part 2 — the outbid filter on the buy screen

`api.AuctionFilter` gains `OutbidOnly bool`; `request.AuctionGetAll` sends
`outbid=true` only when it is set, matching how every other filter is omitted
rather than sent empty.

`auctionfilter.Widget` gains `mvsScope`, a `multivalueselector[Option]` with two
entries:

| Code | Label |
| --- | --- |
| `""` | All auctions |
| `outbid` | My outbid auctions |

It goes **first** in the layout and in the widget's `FocusManager`: it chooses
which set is being searched, and every other control narrows within that set.

Its values are populated in `Init()` before the `Reset()` call, since the labels
are localized and `Init()` runs on every screen entry. `Reset()` returns it to
index 0, so leaving and re-entering the screen never resurrects the previous
visit's scope.

`buildFilter` takes the scope code as its first parameter and sets
`OutbidOnly: scope == "outbid"`.

**`screen/auctionhouse/buy/buy.go` does not change.** Paging, `reload`,
`loadMore`, `afterAction`, `reportCount` and the applied-versus-edited filter
invariant all work on the new filter unmodified. That is the point of expressing
this as a filter.

One consequence, accepted: bidding from an outbid-filtered list makes the row
disappear on `afterAction`'s reload, because the player now leads it. The status
line reports "Bid placed" while the row leaves. Correct, and quiet.

## Part 3 — the manage screen

```
manage.Screen
├── title, characterInfo
├── HBoxFixedRatioLayout 0.5 / 0.5
│   ├── VBox: label + widgetlist[AuctionResponse]   own listings, focusable
│   └── VBox: label + widgetlist[AuctionResponse]   winning bids, focusable
├── statusMessage
└── help
```

`widgetlist` has no title of its own, so each pane's header is a `label.Widget`
above it carrying the pane name and its row count.

Both lists use `auctionlistitem.Constructor` and `SetFilterable(false)`, as the
buy screen's does.

| Pane | Source | Actions |
| --- | --- | --- |
| Your auctions | `GET /auction/own` | `i` inspect, `c` cancel |
| Your winning bids | `GET /auction/bids` | `i` inspect |

### Flows

- `OnEnter` — switch to `keybind.ContextAuctionManage`, refresh character info,
  fetch both lists, focus the own-listings pane. As on the buy screen, the count
  report is gated on the earlier calls having succeeded so it cannot overwrite
  their error.
- Tab / Shift+Tab — between the two panes, through the screen's `FocusManager`.
- `i` — open `auctioninspect` for the focused pane's selected row. Read-only, so
  it is served by both panes.
- `c` — own pane only. A `popup.NewYesNo` naming the item and quantity, then
  `POST /auction/cancel`, then reload both lists and the character info. The
  items come back by mail, so the success message follows `directBuy`'s shape
  and says so.
- `r` — refetch both lists.
- Esc — back to the auction house menu.

`updateFocusKeybinds` hides `c` unless the own-listings pane holds focus,
following the rule established in be1b763: the help line offers only what the
focused pane will actually do. Like the buy screen, `Update` wraps the real
handler so it runs on every path, including dialog exits, which reset every
binding to visible.

Cancelling is the one destructive action here, so it is confirmed rather than
immediate, and the confirmation names what is being cancelled.

### Wiring

- `btManageOnClicked` returns `orvyn.SwitchScreen(screen.IDAuctionHouseManage)`.
- `menu.Update` gains the missing `MKey` case beside `SKey` and `BKey`.
- The screen is registered in `main.go` beside the other auction house screens.
- `keybind.ContextAuctionManage` is registered in `internal/keybind/context.go`.

## Errors

- Every call goes through `helper.Fetch` / `helper.SendRequest`; failures land
  in `statusMessage.SetError`.
- An empty pane is a status line, not an error. A player with no listings and a
  failed request must not look alike.
- A partial load — own listings fetched, bids failed — shows the error and
  leaves the other pane populated. Neither pane depends on the other.
- `auctioninspect` switches the help context on entry and never switches back,
  so the manage screen restores it on `DialogExitMsg`, cancel path included.
- 401 handling is untouched: a messageless 401 expires the session, a business
  401 ("no auction house at this location") carries its message.

## Testing

Server:

- `RecordBid` — inserted once, and a second bid on the same auction is a no-op.
- The `OutbidFor` filter — an auction the character leads is excluded, one they
  bid on and lost is included, one they never touched is excluded, and a settled
  one is excluded. `Total` matches the rows returned.
- `parseAuctionFilter` — `outbid=true`, `outbid=false`, absent, and a malformed
  value producing a 400.

Client, only where the logic is pure — widget constructors need a theme, so they
are not unit-tested, consistent with the rest of the repo:

- `buildFilter` — the scope code set and unset, in combination with the existing
  filters.
- `core/request/auction_test.go` — `outbid` sent when set, omitted when not.

Everything else is verified with `go build ./...`, `go vet ./...`, `gofmt -l`
and `go test ./...` in both repos. The screens themselves need a TTY and a
running backend, so they are not exercised automatically; that gap is stated
rather than papered over.

## Implementation order

Part 1 stands alone and unblocks the other two. Part 2 needs Part 1's endpoint
behaviour. Part 3 depends on neither and can be built in parallel with Part 2 —
it consumes only endpoints that already exist.

## Conventions

Comments explain why, not what, and stay short. No comment restates a line of
code below it.
