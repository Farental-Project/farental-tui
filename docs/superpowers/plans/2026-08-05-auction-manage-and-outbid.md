# Auction manage screen and outbid tracking — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let a player find the still-running auctions they were outbid on, and give the auction house a manage screen for their own listings and winning bids.

**Architecture:** The server gains an `auction_bids` participation table written inside the bid transaction, and exposes "I was outbid here" as a new `outbid` parameter on the existing `/auction/all` browse endpoint rather than as a new endpoint — so paging, `Total` and every existing filter keep working untouched. The client adds one scope control to the auction filter panel, and a new manage screen consuming the two endpoints that already exist.

**Tech Stack:** Go, gorm, fiber, MariaDB (server); Go, bubbletea, orvyn, lokyn (client).

## Global Constraints

- Design spec: `docs/superpowers/specs/2026-08-05-auction-manage-and-outbid-design.md` in the client repo. Read it before starting.
- Two repositories:
  - Server: `/home/halsten/Dev/Farental/farental-cli`, Go module rooted at `src/`. Currently on `main`, clean.
  - Client: `/home/halsten/Dev/Farental/farental-tui`, Go module rooted at `src/`. Currently on branch `auction-manage-screen`.
- Tasks 1–4 are server-only. Tasks 5–9 are client-only. Never edit one repo from a task belonging to the other.
- Server tests run against a live MariaDB test database via `testutil`. They work in this environment — a full `go test ./...` from `farental-cli/src` takes a few minutes.
- Client verification is `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test ./...`, all from `farental-tui/src`. The screens need a TTY and a running backend, so they are not exercised automatically.
- Comments explain **why**, not what, and stay short. No comment restates the line of code below it.
- Commit messages are Conventional Commits with a lowercase subject, matching both repos' history (`feat: highlight auction prices and split them from the listing metadata`).
- Every user-facing client string is wrapped in `lokyn.L(...)`. No new server-side strings are introduced by this plan, so the `createmissingstrints` pipeline is not involved.

---

## Task 1: The participation record

Creates the table, registers it for AutoMigrate, puts it in the migration, and adds the repository method that writes it.

**Files:**
- Modify: `farental-cli/src/model/data/auction.go` (append after the `Auction` struct)
- Modify: `farental-cli/src/config/database.go:329-331`
- Modify: `farental-cli/src/repository/auction.go` (append)
- Modify: `farental-cli/migrations/structure/s1.5.0__auctions.sql`
- Test: `farental-cli/src/repository/auctionbid_test.go` (create)

**Interfaces:**
- Consumes: nothing.
- Produces: `data.AuctionBid{AuctionID, CharacterID uint}`; `(*repository.AuctionRepository).RecordBid(auctionID, characterID uint) error`.

- [ ] **Step 1: Branch the server repo**

```bash
cd /home/halsten/Dev/Farental/farental-cli
git checkout -b auction-outbid-tracking
```

- [ ] **Step 2: Write the failing test**

Create `farental-cli/src/repository/auctionbid_test.go`.

`seedFilterItem` and `seedFilterAuction` already exist in `auction_test.go` in this same package (`repository_test`) and register their own `t.Cleanup`.

```go
package repository_test

import (
	"farental/model/data"
	"farental/repository"
	"farental/testutil"
	"testing"
)

// A player may bid on the same auction more than once — outbid, then bidding
// again. The participation row must not multiply, and nothing downstream
// should have to deduplicate it.
func TestRecordBidIsIdempotent(t *testing.T) {
	repo := &repository.AuctionRepository{DB: testutil.DB}

	item := seedFilterItem(t, nil)
	auction := seedFilterAuction(t, item)
	characterID := testutil.GetCharacter().ID

	if err := repo.RecordBid(auction.ID, characterID); err != nil {
		t.Fatalf("first RecordBid: %v", err)
	}

	if err := repo.RecordBid(auction.ID, characterID); err != nil {
		t.Fatalf("second RecordBid: %v", err)
	}

	if got := auctionBidCount(t, auction.ID, characterID); got != 1 {
		t.Errorf("participation rows = %d, want 1", got)
	}
}

// The rows are swept by deleting the auction they belong to (kic), so the
// cascade is load-bearing rather than incidental.
func TestAuctionBidCascadesWithItsAuction(t *testing.T) {
	repo := &repository.AuctionRepository{DB: testutil.DB}

	item := seedFilterItem(t, nil)
	auction := seedFilterAuction(t, item)
	characterID := testutil.GetCharacter().ID

	if err := repo.RecordBid(auction.ID, characterID); err != nil {
		t.Fatalf("RecordBid: %v", err)
	}

	if err := testutil.DB.Unscoped().Delete(&data.Auction{}, auction.ID).Error; err != nil {
		t.Fatalf("delete auction: %v", err)
	}

	if got := auctionBidCount(t, auction.ID, characterID); got != 0 {
		t.Errorf("participation rows after deleting the auction = %d, want 0", got)
	}
}

func auctionBidCount(t *testing.T, auctionID, characterID uint) int64 {
	t.Helper()

	var count int64

	err := testutil.DB.Model(&data.AuctionBid{}).
		Where("auction_id = ? AND character_id = ?", auctionID, characterID).
		Count(&count).Error

	if err != nil {
		t.Fatalf("count participation rows: %v", err)
	}

	return count
}
```

- [ ] **Step 3: Run the test to verify it fails**

```bash
cd /home/halsten/Dev/Farental/farental-cli/src
go test ./repository/ -run 'TestRecordBid|TestAuctionBidCascades' -v
```

Expected: build failure — `undefined: data.AuctionBid` and `repo.RecordBid undefined`.

- [ ] **Step 4: Add the model**

Append to `farental-cli/src/model/data/auction.go`:

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

The composite primary key is the uniqueness constraint; there is no surrogate id.

- [ ] **Step 5: Register it for AutoMigrate**

In `farental-cli/src/config/database.go`, directly after the existing `// Auction` block (lines 329-332) and before `// Task`:

```go
	// Auction bid
	if err = db.AutoMigrate(&data.AuctionBid{}); err != nil {
		return
	}
```

AutoMigrate takes an explicit model list rather than discovering models, and it is the only thing that creates tables in development.

- [ ] **Step 6: Add RecordBid**

Append to `farental-cli/src/repository/auction.go` (the `clause` import is new):

```go
// RecordBid notes that the character has bid on the auction, so they can still
// be found once someone outbids them and the auction row stops referencing them.
//
// Bidding again on the same auction is a no-op: the composite primary key
// rejects the duplicate and OnConflict swallows it, so no caller has to know
// whether this is a first bid or a fifth.
func (r *AuctionRepository) RecordBid(auctionID, characterID uint) error {
	return r.DB.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&data.AuctionBid{
			AuctionID:   auctionID,
			CharacterID: characterID,
		}).Error
}
```

Add `"gorm.io/gorm/clause"` to the import block.

- [ ] **Step 7: Run the tests to verify they pass**

```bash
cd /home/halsten/Dev/Farental/farental-cli/src
go test ./repository/ -run 'TestRecordBid|TestAuctionBidCascades' -v
```

Expected: PASS, both tests. If the cascade test fails, the development database predates the model — drop the `auction_bids` table and let AutoMigrate recreate it.

- [ ] **Step 8: Add the table to the migration**

In `farental-cli/migrations/structure/s1.5.0__auctions.sql`:

1. Extend the header `Description` to mention `auction_bids` alongside `auctions` and the users column.
2. Append this block after the `auctions` CREATE TABLE, before the `═══` line that closes the migration body:

```sql
-- ─── auction_bids ────────────────────────────────────────────────────────────
-- One row per (auction, character) that has ever bid, with no amount: an outbid
-- player was already refunded by mail, and the current bid lives on the auction.
-- It exists so a player can still be matched to an auction after being outbid,
-- which `auctions` alone cannot express — current_bidder_id names one character.
--
-- Both foreign keys cascade, unlike the ones on `auctions` above. A
-- participation row holds no goods and no standing bid, so it must never block
-- deleting the auction it belongs to or the character that wrote it; the kic
-- engine's auction purge relies on the first of those.
CREATE TABLE IF NOT EXISTS `auction_bids` (
  `auction_id`   bigint(20) unsigned NOT NULL,
  `character_id` bigint(20) unsigned NOT NULL,
  PRIMARY KEY (`auction_id`,`character_id`),
  KEY `idx_auction_bids_character` (`character_id`),
  CONSTRAINT `fk_auction_bids_auction` FOREIGN KEY (`auction_id`)
    REFERENCES `auctions` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_auction_bids_character` FOREIGN KEY (`character_id`)
    REFERENCES `characters` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;
```

3. Add `SHOW CREATE TABLE auction_bids;` to the post-conditions list in the footer.

`s1.5.0` is edited rather than superseded by an `s1.6.0`: it is not in production yet, the guard still requires structure `1.4.0`, and prod will receive both tables in one pass.

- [ ] **Step 9: Verify the DDL matches what AutoMigrate produces**

```bash
cd /home/halsten/Dev/Farental/farental-cli/src
go test ./repository/ -run TestRecordBid   # ensures AutoMigrate has run
```

Then dump the live DDL. The test database's credentials live in `farental-cli/src/.env`; it cannot be `source`d (line 11 contains an `&` that zsh rejects), so read the four values out of it:

```bash
cd /home/halsten/Dev/Farental/farental-cli/src
DB_NAME=$(grep '^DB_NAME=' .env | cut -d= -f2-)
DB_USERNAME=$(grep '^DB_USERNAME=' .env | cut -d= -f2-)
DB_PASSWORD=$(grep '^DB_PASSWORD=' .env | cut -d= -f2-)
DB_HOST=$(grep '^DB_URL=' .env | cut -d= -f2- | cut -d: -f1)
mysql -u "$DB_USERNAME" -p"$DB_PASSWORD" -h "$DB_HOST" "$DB_NAME" \
  -e 'SHOW CREATE TABLE auction_bids\G'
```

Column order, types, key names and the two constraints must match the SQL above. The migration file's own header requires this: a database built from the models and one built from the migration have to be identical. Adjust the SQL to match the live output if they differ.

- [ ] **Step 10: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-cli
git add src/model/data/auction.go src/config/database.go src/repository/auction.go \
        src/repository/auctionbid_test.go migrations/structure/s1.5.0__auctions.sql
git commit -m "feat: record which characters have bid on an auction"
```

---

## Task 2: Record a bid as it is placed

**Files:**
- Modify: `farental-cli/src/serverapi/controller/auctiontrade.go:108-125`
- Test: `farental-cli/src/serverapi/controller/auctionbid_test.go` (create)

**Interfaces:**
- Consumes: `(*repository.AuctionRepository).RecordBid(auctionID, characterID uint) error` from Task 1.
- Produces: a participation row for every accepted bid. Nothing new is exported.

- [ ] **Step 1: Write the failing test**

Create `farental-cli/src/serverapi/controller/auctionbid_test.go`. The helpers used here — `loadAuctionCharacter`, `secondCharacter`, `seedAuctionItem`, `seedAuction`, `setMoney`, `setupAuctionApp`, `postJSON` — all already exist in `auction_test.go` in this same package.

```go
package controller

import (
	"testing"
	"time"

	"farental/model/api"
	"farental/model/data"
	"farental/testutil"
)

// The participation row is what makes a player findable once someone outbids
// them, so it has to be written by the bid itself rather than by anything the
// client remembers to send afterwards.
func TestAuctionBidRecordsParticipation(t *testing.T) {
	bidder := loadAuctionCharacter(t, testutil.GetCharacter().ID)
	seller := secondCharacter(t)
	item := seedAuctionItem(t, true, 10)

	auction := seedAuction(t, seller.ID, item, 1, 100, 0, nil, time.Now().Add(time.Hour))

	setMoney(t, bidder, 1000)

	app := setupAuctionApp(t, bidder)

	resp := postJSON(t, app, "/auction/bid", api.AuctionBidBody{ID: auction.ID, Bid: 150})

	if resp.Code != 200 {
		t.Fatalf("bid status = %d, want 200", resp.Code)
	}

	var count int64

	err := testutil.DB.Model(&data.AuctionBid{}).
		Where("auction_id = ? AND character_id = ?", auction.ID, bidder.ID).
		Count(&count).Error

	if err != nil {
		t.Fatalf("count participation rows: %v", err)
	}

	if count != 1 {
		t.Errorf("participation rows = %d, want 1", count)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

```bash
cd /home/halsten/Dev/Farental/farental-cli/src
go test ./serverapi/controller/ -run TestAuctionBidRecordsParticipation -v
```

Expected: FAIL with `participation rows = 0, want 1`.

- [ ] **Step 3: Record the bid inside the transaction**

In `farental-cli/src/serverapi/controller/auctiontrade.go`, in `AuctionMakeBid`, between the `data.SaveWithoutAssociations` error check and the `eventlog.CharacterLogAddEntry` call:

```go
	// Inside the transaction for the same reason the log entry below is: a
	// rolled-back bid must not leave a participation row claiming the player
	// bid here, or the outbid list offers an auction they never touched.
	//
	// A failure rolls the bid back rather than committing an untracked one —
	// the money has already moved, and the transaction is the only thing that
	// can put both back.
	if err = ctx.Auctions.RecordBid(auctionRecord.ID, ctx.Character.ID); err != nil {
		slog.Error("auction make bid while recording participation", "err", err,
			"auctionID", auctionRecord.ID)
		ctx.DB.Rollback()
		return c.SendStatus(fiber.StatusInternalServerError)
	}
```

- [ ] **Step 4: Run the test to verify it passes**

```bash
cd /home/halsten/Dev/Farental/farental-cli/src
go test ./serverapi/controller/ -run 'TestAuctionBid' -v
```

Expected: PASS, including the pre-existing `TestAuctionBidCanMeetAskingPrice` and `TestAuctionBidRejectsRaisingYourOwnBid`.

- [ ] **Step 5: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-cli
git add src/serverapi/controller/auctiontrade.go src/serverapi/controller/auctionbid_test.go
git commit -m "feat: record participation when a bid is placed"
```

---

## Task 3: The outbid browse filter

**Files:**
- Modify: `farental-cli/src/repository/auction.go` — `AuctionFilter` struct, and the `scope` closure inside `FindAllWithPreloads`
- Modify: `farental-cli/src/serverapi/controller/auction.go` — `parseAuctionFilter` and the `AuctionGetAll` swagger block
- Test: `farental-cli/src/serverapi/controller/auctionbid_test.go` (append)

**Interfaces:**
- Consumes: `data.AuctionBid` from Task 1.
- Produces: `repository.AuctionFilter.OutbidFor uint`; `GET /auction/all?outbid=true`.

- [ ] **Step 1: Write the failing test**

Append to `farental-cli/src/serverapi/controller/auctionbid_test.go`. Add `"encoding/json"`, `"net/http/httptest"` and `"farental/repository"` to its imports.

```go
// The whole point of the participation table: an auction the player bid on and
// then lost the lead in is the one they want back. One they still lead needs no
// action, and one they never touched is not theirs to track.
func TestAuctionListOutbidFilter(t *testing.T) {
	browser := loadAuctionCharacter(t, testutil.GetCharacter().ID)
	seller := secondCharacter(t)
	rival := secondCharacter(t)
	item := seedAuctionItem(t, true, 10)

	repo := &repository.AuctionRepository{DB: testutil.DB}

	// Bid on, then outbid by the rival — the one row the filter must return.
	outbid := seedAuction(t, seller.ID, item, 1, 200, 0, &rival.ID, time.Now().Add(time.Hour))

	if err := repo.RecordBid(outbid.ID, browser.ID); err != nil {
		t.Fatalf("record bid on the outbid auction: %v", err)
	}

	// Bid on and still leading — nothing to do about it.
	leading := seedAuction(t, seller.ID, item, 1, 200, 0, &browser.ID, time.Now().Add(time.Hour))

	if err := repo.RecordBid(leading.ID, browser.ID); err != nil {
		t.Fatalf("record bid on the led auction: %v", err)
	}

	// Never bid on.
	untouched := seedAuction(t, seller.ID, item, 1, 100, 0, nil, time.Now().Add(time.Hour))

	app := setupAuctionApp(t, browser)

	req := httptest.NewRequest("GET", "/auction/all?outbid=true", nil)
	resp, err := app.Test(req, -1)

	if err != nil {
		t.Fatalf("request: %v", err)
	}

	var listed api.AuctionListResponse

	if err := json.NewDecoder(resp.Body).Decode(&listed); err != nil {
		t.Fatalf("decode: %v", err)
	}

	got := map[uint]bool{}

	for _, a := range listed.Auctions {
		got[a.ID] = true
	}

	if !got[outbid.ID] {
		t.Errorf("outbid auction %d missing from the filtered list", outbid.ID)
	}

	if got[leading.ID] {
		t.Errorf("auction %d is still led by the browser and must not be listed", leading.ID)
	}

	if got[untouched.ID] {
		t.Errorf("auction %d was never bid on and must not be listed", untouched.ID)
	}

	// Total drives paging, so it has to agree with the rows or later pages
	// promise auctions no page contains.
	if listed.Total != int64(len(listed.Auctions)) {
		t.Errorf("Total = %d but %d auctions returned", listed.Total, len(listed.Auctions))
	}

	// outbid=false is not "show me the ones I lead" — it is no narrowing at
	// all, the same as leaving the parameter out.
	req = httptest.NewRequest("GET", "/auction/all?outbid=false", nil)
	resp, err = app.Test(req, -1)

	if err != nil {
		t.Fatalf("request with outbid=false: %v", err)
	}

	var unfiltered api.AuctionListResponse

	if err := json.NewDecoder(resp.Body).Decode(&unfiltered); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if unfiltered.Total <= listed.Total {
		t.Errorf("outbid=false Total = %d, want more than the %d the filter returned",
			unfiltered.Total, listed.Total)
	}
}

// QueryBool would map a malformed value to false and answer with the entire
// unfiltered house, which is the failure the minStat parsing already avoids.
func TestAuctionListRejectsMalformedOutbid(t *testing.T) {
	browser := loadAuctionCharacter(t, testutil.GetCharacter().ID)
	app := setupAuctionApp(t, browser)

	req := httptest.NewRequest("GET", "/auction/all?outbid=perhaps", nil)
	resp, err := app.Test(req, -1)

	if err != nil {
		t.Fatalf("request: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

```bash
cd /home/halsten/Dev/Farental/farental-cli/src
go test ./serverapi/controller/ -run 'TestAuctionListOutbidFilter|TestAuctionListRejectsMalformedOutbid' -v
```

Expected: both FAIL — the parameter is ignored today, so the first sees the led and untouched auctions listed and the second gets 200.

- [ ] **Step 3: Add the filter field**

In `farental-cli/src/repository/auction.go`, inside the `AuctionFilter` struct, after `ExcludeSellerID`:

```go
	// OutbidFor narrows to the auctions this character has bid on and no longer
	// leads. Zero leaves the browse unnarrowed.
	OutbidFor uint
```

- [ ] **Step 4: Apply it in the scope closure**

In `FindAllWithPreloads`, inside the `scope` closure, after the `ExcludeSellerID` block and before the `itemScope` block:

```go
		if f.OutbidFor != 0 {
			bids := r.DB.Model(&data.AuctionBid{}).
				Select("auction_id").
				Where("character_id = ?", f.OutbidFor)

			// current_bidder_id is never NULL on a row reached through a bid:
			// the bid that wrote the participation row set it, and nothing
			// clears the leader while the auction is Active — a later bidder
			// replaces them, and every other outcome sets a terminal status.
			db = db.Where("id IN (?)", bids).
				Where("current_bidder_id <> ?", f.OutbidFor)
		}
```

It belongs in `scope` rather than `itemScope` because it narrows the auction, not the item — and because `scope` is applied to both the `Count` and the page query, which is what keeps `Total` honest.

- [ ] **Step 5: Parse the query parameter**

In `farental-cli/src/serverapi/controller/auction.go`, in `parseAuctionFilter`, after the `weaponType` block and before the `statCode`/`skillCode` handling:

```go
	// strconv.ParseBool rather than c.QueryBool, for the reason spelled out
	// above minStat: QueryBool maps an unparsable value to its default, and
	// here that default answers a malformed request with the whole house.
	if raw := c.Query("outbid", ""); raw != "" {
		value, err := strconv.ParseBool(raw)

		if err != nil {
			invalid("outbid", raw)
		} else if value {
			filter.OutbidFor = ctx.Character.ID
		}
	}
```

`strconv` is already imported by this file.

- [ ] **Step 6: Document the parameter**

In the swagger block above `AuctionGetAll`, after the `minStat` line:

```go
// @Param			outbid		query	boolean	false	"Only auctions the current character has bid on and no longer leads"
```

and append this sentence to the existing `@Description` line, after "minStat is ignored unless stat or skill is also given.":

```
outbid=true narrows to the auctions the current character has bid on and no longer leads.
```

- [ ] **Step 7: Run the tests to verify they pass**

```bash
cd /home/halsten/Dev/Farental/farental-cli/src
go test ./serverapi/controller/ ./repository/ -run 'Auction' -v
```

Expected: PASS, including the pre-existing `TestAuctionListIsPagedAndFilterable` and `TestAuctionListExcludesOwnAuctions`.

- [ ] **Step 8: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-cli
git add src/repository/auction.go src/serverapi/controller/auction.go \
        src/serverapi/controller/auctionbid_test.go
git commit -m "feat: filter the auction browse to the outbid ones"
```

---

## Task 4: kic purges settled auctions

**Files:**
- Modify: `farental-cli/src/serverapi/system/kic/keepitclean.go`
- Test: `farental-cli/src/serverapi/system/kic/auction_test.go` (create)

**Interfaces:**
- Consumes: `data.AuctionBid` from Task 1.
- Produces: `cleanOldAuctions(db *gorm.DB) (int64, error)` — unexported, used by `Engine` and its test.

- [ ] **Step 1: Write the failing test**

Create `farental-cli/src/serverapi/system/kic/auction_test.go`. This package's `TestMain` already calls `testutil.Setup()`, `InitUser()` and `InitCharacter()`, so a character exists.

```go
package kic

import (
	"testing"
	"time"

	"farental/model/data"
	"farental/testutil"
)

// seedKicItem creates the item an auction needs. STName is a tag, not a
// sentence; nothing in this test renders it.
func seedKicItem(t *testing.T) *data.Item {
	t.Helper()

	item := data.Item{
		STName:        data.StrintTag("items.kic_test_item"),
		STDescription: data.StrintTag("items.kic_test_item_desc"),
		IsShareable:   true,
		IsSellable:    true,
		MaxStackCount: 10,
	}

	if err := testutil.DB.Create(&item).Error; err != nil {
		t.Fatalf("seed item: %v", err)
	}

	t.Cleanup(func() {
		testutil.DB.Unscoped().Delete(&data.Item{}, item.ID)
	})

	return &item
}

func seedKicAuction(t *testing.T, status data.AuctionStatus, end time.Time) *data.Auction {
	t.Helper()

	auction := data.Auction{
		Status:       status,
		ItemID:       seedKicItem(t).ID,
		Quantity:     1,
		CurrentBid:   100,
		SellerID:     testutil.GetCharacter().ID,
		Duration:     data.AuctionDurationShort,
		EndTimestamp: end,
	}

	if err := testutil.DB.Create(&auction).Error; err != nil {
		t.Fatalf("seed auction: %v", err)
	}

	t.Cleanup(func() {
		testutil.DB.Unscoped().Delete(&data.Auction{}, auction.ID)
	})

	return &auction
}

func auctionExists(t *testing.T, id uint) bool {
	t.Helper()

	var count int64

	if err := testutil.DB.Model(&data.Auction{}).Where("id = ?", id).Count(&count).Error; err != nil {
		t.Fatalf("count auctions: %v", err)
	}

	return count > 0
}

// Settled auctions were never swept before this, so they accumulated forever.
// A live one must survive regardless of age, and an unsettled expired one is
// the settlement engine's to deal with — deleting it would strand the goods.
func TestCleanOldAuctions(t *testing.T) {
	old := time.Now().AddDate(0, 0, -40)
	recent := time.Now().AddDate(0, 0, -2)

	oldSold := seedKicAuction(t, data.AuctionStatusSold, old)
	oldExpired := seedKicAuction(t, data.AuctionStatusExpired, old)
	recentSold := seedKicAuction(t, data.AuctionStatusSold, recent)
	live := seedKicAuction(t, data.AuctionStatusActive, time.Now().Add(time.Hour))
	unsettled := seedKicAuction(t, data.AuctionStatusActive, old)

	bid := data.AuctionBid{
		AuctionID:   oldSold.ID,
		CharacterID: testutil.GetCharacter().ID,
	}

	if err := testutil.DB.Create(&bid).Error; err != nil {
		t.Fatalf("seed participation row: %v", err)
	}

	if _, err := cleanOldAuctions(testutil.DB); err != nil {
		t.Fatalf("cleanOldAuctions: %v", err)
	}

	if auctionExists(t, oldSold.ID) {
		t.Error("a sold auction older than the window survived")
	}

	if auctionExists(t, oldExpired.ID) {
		t.Error("an expired auction older than the window survived")
	}

	if !auctionExists(t, recentSold.ID) {
		t.Error("a sold auction inside the window was deleted")
	}

	if !auctionExists(t, live.ID) {
		t.Error("a live auction was deleted")
	}

	if !auctionExists(t, unsettled.ID) {
		t.Error("an expired but unsettled auction was deleted before settlement")
	}

	var bids int64

	err := testutil.DB.Model(&data.AuctionBid{}).
		Where("auction_id = ?", oldSold.ID).Count(&bids).Error

	if err != nil {
		t.Fatalf("count participation rows: %v", err)
	}

	if bids != 0 {
		t.Errorf("participation rows left behind = %d, want 0 — the cascade did not fire", bids)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

```bash
cd /home/halsten/Dev/Farental/farental-cli/src
go test ./serverapi/system/kic/ -run TestCleanOldAuctions -v
```

Expected: build failure — `undefined: cleanOldAuctions`.

- [ ] **Step 3: Add the sweep function**

Append to `farental-cli/src/serverapi/system/kic/keepitclean.go`, beside `cleanOrphanPendingFightResults`:

```go
// cleanOldAuctions deletes settled auctions past the retention window. Their
// auction_bids rows go with them through ON DELETE CASCADE.
//
// Active is excluded rather than filtered by date: an expired auction that is
// still Active has not been settled, so its goods and the standing bid are
// still owed to somebody and the settlement engine has to see it.
//
// A cancelled auction's end_timestamp is still in the future, so it outlives
// the window by up to its remaining duration. Harmless at 96h worst case, and
// the alternative is a settled_timestamp column nothing else needs.
func cleanOldAuctions(db *gorm.DB) (int64, error) {
	limitDate := time.Now().AddDate(0, 0, -30)

	var auctions []data.Auction

	result := db.Where("status <> ? AND end_timestamp < ?",
		data.AuctionStatusActive, limitDate).Find(&auctions)

	if result.Error != nil || result.RowsAffected == 0 {
		return 0, result.Error
	}

	if err := db.Unscoped().Delete(&auctions).Error; err != nil {
		return 0, err
	}

	return result.RowsAffected, nil
}
```

- [ ] **Step 4: Call it from the engine loop**

In `Engine`, after the `CHAT MESSAGES` block and before the closing brace of the `for` loop:

```go
		// AUCTIONS
		// Delete old settled auctions (30 days), and the participation rows
		// that cascade off them.
		auctionCount, err := cleanOldAuctions(db)

		if err != nil {
			slog.Error("kic engine: clean old auctions failed", "err", err)
		}

		logStep(auctionCount, "Auctions")
```

`:=` is correct here even though `err` already exists in this scope — `auctionCount` is new, which is all a short declaration needs.

- [ ] **Step 5: Run the test to verify it passes**

```bash
cd /home/halsten/Dev/Farental/farental-cli/src
go test ./serverapi/system/kic/ -v
```

Expected: PASS, including the pre-existing `TestCleanOrphanPendingFightResults`.

- [ ] **Step 6: Run the whole server suite**

```bash
cd /home/halsten/Dev/Farental/farental-cli/src
go build ./... && go vet ./... && gofmt -l . && go test ./...
```

Expected: no build errors, no vet findings, no files listed by gofmt, all packages ok.

- [ ] **Step 7: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-cli
git add src/serverapi/system/kic/keepitclean.go src/serverapi/system/kic/auction_test.go
git commit -m "feat: sweep settled auctions in the kic engine"
```

---

## Task 5: Send the outbid parameter from the client

**Files:**
- Modify: `farental-tui/src/core/data/api/auction.go:169-189`
- Modify: `farental-tui/src/core/request/auction.go:14-53`
- Test: `farental-tui/src/core/request/auction_test.go`

**Interfaces:**
- Consumes: `GET /auction/all?outbid=true` from Task 3.
- Produces: `api.AuctionFilter.OutbidOnly bool`, sent as `outbid=true` by `request.AuctionGetAll`.

- [ ] **Step 1: Write the failing test**

Append to `farental-tui/src/core/request/auction_test.go`:

```go
// Being outbid is a property of the auction, so it rides the browse endpoint
// as a filter rather than needing an endpoint of its own — but like every
// other filter it must be absent, not empty, when it is not asked for.
func TestAuctionGetAllSendsOutbidOnlyWhenSet(t *testing.T) {
	Init(resty.New())

	params := AuctionGetAll(1, api.AuctionFilter{OutbidOnly: true}).QueryParam

	if got := params.Get("outbid"); got != "true" {
		t.Errorf("outbid = %q, want %q", got, "true")
	}

	params = AuctionGetAll(1, api.AuctionFilter{}).QueryParam

	if _, ok := params["outbid"]; ok {
		t.Error("outbid sent for a filter that did not ask for it, want it omitted")
	}
}
```

Also add `"outbid"` to the omitted-name list in the existing `TestAuctionGetAllOmitsUnsetFilters`.

- [ ] **Step 2: Run the test to verify it fails**

```bash
cd /home/halsten/Dev/Farental/farental-tui/src
go test ./core/request/ -run TestAuctionGetAll -v
```

Expected: build failure — `unknown field OutbidOnly in struct literal`.

- [ ] **Step 3: Add the field**

In `farental-tui/src/core/data/api/auction.go`, inside `AuctionFilter`, after `HasMinStat`:

```go
	// OutbidOnly narrows to the auctions this character has bid on and no
	// longer leads. The server resolves whose auctions those are from the
	// session, so the flag carries no character.
	OutbidOnly bool
```

- [ ] **Step 4: Send it**

In `farental-tui/src/core/request/auction.go`, inside `AuctionGetAll`, after the `minStat` block:

```go
	if filter.OutbidOnly {
		r.SetQueryParam("outbid", "true")
	}
```

- [ ] **Step 5: Run the tests to verify they pass**

```bash
cd /home/halsten/Dev/Farental/farental-tui/src
go test ./core/request/ -v
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/core/data/api/auction.go src/core/request/auction.go src/core/request/auction_test.go
git commit -m "feat: send the outbid browse filter"
```

---

## Task 6: The scope control in the filter panel

**Files:**
- Modify: `farental-tui/src/widget/auctionfilter/filter.go`
- Modify: `farental-tui/src/widget/auctionfilter/auctionfilter.go`
- Test: `farental-tui/src/widget/auctionfilter/filter_test.go`

**Interfaces:**
- Consumes: `api.AuctionFilter.OutbidOnly` from Task 5.
- Produces: `auctionfilter.Widget.GetFilter()` returns a filter whose `OutbidOnly` reflects the scope selector. The screen-facing surface (`Init`, `Reset`, `GetFilter`, `SetOptions`, `AppliedMsg`) is unchanged, so `screen/auctionhouse/buy` needs no edit.

**Constant:** the scope code for outbid-only is the string `"outbid"`, defined once in `filter.go` as `scopeOutbid`.

- [ ] **Step 1: Write the failing test**

Append to `farental-tui/src/widget/auctionfilter/filter_test.go`:

```go
func TestBuildFilterScope(t *testing.T) {
	cases := []struct {
		name  string
		scope string
		want  bool
	}{
		{"empty scope browses the whole house", "", false},
		{"outbid scope narrows to the player's lost leads", "outbid", true},
	}

	for _, tc := range cases {
		f := buildFilter(tc.scope, "", "", "", "", "", "")

		if f.OutbidOnly != tc.want {
			t.Errorf("%s: OutbidOnly = %v, want %v", tc.name, f.OutbidOnly, tc.want)
		}
	}
}

// The scope narrows which auctions are searched; the rest narrow within them.
// Nothing about the two is exclusive, so both have to survive one call.
func TestBuildFilterScopeCombinesWithItemFilters(t *testing.T) {
	f := buildFilter("outbid", "equipment", "hea", "sword", "str", "", "5")

	if !f.OutbidOnly {
		t.Error("OutbidOnly = false, want true")
	}

	if f.Kind != api.AuctionItemKindEquipment {
		t.Errorf("Kind = %q, want %q", f.Kind, api.AuctionItemKindEquipment)
	}

	if f.SlotCode != "hea" {
		t.Errorf("SlotCode = %q, want %q", f.SlotCode, "hea")
	}

	if !f.HasMinStat || f.MinStat != 5 {
		t.Errorf("MinStat = %d (set %v), want 5 (set true)", f.MinStat, f.HasMinStat)
	}
}
```

The five existing calls in this file take six arguments and will stop compiling. Give each an empty scope as its new first argument — the scope is not part of what they cover:

| Line | From | To |
| --- | --- | --- |
| 9 | `buildFilter("", "", "", "", "", "")` | `buildFilter("", "", "", "", "", "", "")` |
| 17 | `buildFilter("equipment", "hea", "sword", "str", "", "5")` | `buildFilter("", "equipment", "hea", "sword", "str", "", "5")` |
| 36 | `buildFilter("", "", "", "", "", "5")` | `buildFilter("", "", "", "", "", "", "5")` |
| 44 | `buildFilter("", "", "", "", "swordsmanship", "-5")` | `buildFilter("", "", "", "", "", "swordsmanship", "-5")` |
| 56 | `buildFilter("", "", "", "str", "", "-")` | `buildFilter("", "", "", "", "str", "", "-")` |

- [ ] **Step 2: Run the tests to verify they fail**

```bash
cd /home/halsten/Dev/Farental/farental-tui/src
go test ./widget/auctionfilter/ -v
```

Expected: build failure — `too many arguments in call to buildFilter`.

- [ ] **Step 3: Extend buildFilter**

In `farental-tui/src/widget/auctionfilter/filter.go`:

```go
// scopeOutbid is the scope selector's non-empty code. The empty code browses
// the whole house.
const scopeOutbid = "outbid"

// buildFilter assembles the query from raw control values. Kept free of the
// widgets so the rules the server enforces can be tested directly.
//
// scope comes first because it chooses which set of auctions is searched;
// everything after it narrows within that set.
func buildFilter(scope, kind, slot, weaponType, statCode, skillCode, minStat string) api.AuctionFilter {
	f := api.AuctionFilter{
		OutbidOnly:     scope == scopeOutbid,
		Kind:           api.AuctionItemKind(kind),
		SlotCode:       slot,
		WeaponTypeCode: weaponType,
		StatCode:       statCode,
		SkillCode:      skillCode,
	}
```

The rest of the function is unchanged.

- [ ] **Step 4: Run the tests to verify they pass**

```bash
cd /home/halsten/Dev/Farental/farental-tui/src
go test ./widget/auctionfilter/ -v
```

Expected: PASS.

- [ ] **Step 5: Add the control to the widget**

In `farental-tui/src/widget/auctionfilter/auctionfilter.go`, four edits:

1. Struct — add above `lblKind`:

```go
	lblScope *label.Widget
	mvsScope *multivalueselector.Widget[Option]
```

2. `New()` — construct them before the kind controls, register the selector first in the focus manager, and put both first in the layout:

```go
	w.lblScope = label.New("")

	w.mvsScope = multivalueselector.New[Option]()
	w.mvsScope.OnBlur()
```

```go
	w.focusManager.Add(w.mvsScope)
```
(before the existing `w.focusManager.Add(w.mvsKind)`)

```go
	w.layout = layout.NewMaxWidthVBoxLayout(0,
		w.lblScope,
		w.mvsScope,
		w.lblKind,
		…
```

3. `Init()` — the scope options are client-side strings rather than server vocabulary, and they are localized, so they are filled here on every entry. They must be set **before** `w.Reset()`, which selects index 0:

```go
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

	w.lblScope.SetValue(lokyn.L("Show"))
	w.lblKind.SetValue(lokyn.L("Kind"))
	…
```

4. `Reset()` — add as its first line, and `GetFilter()` — add as its first argument:

```go
	w.mvsScope.SetSelected(0)
```

```go
	return buildFilter(
		w.mvsScope.GetSelectedValue().Code,
		w.mvsKind.GetSelectedValue().Code,
		…
```

- [ ] **Step 6: Verify the client builds clean**

```bash
cd /home/halsten/Dev/Farental/farental-tui/src
go build ./... && go vet ./... && gofmt -l . && go test ./...
```

Expected: no output from `gofmt -l`, all packages ok. `screen/auctionhouse/buy` must compile untouched — if it needed an edit, the widget's public surface changed and that is a mistake.

- [ ] **Step 7: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/widget/auctionfilter/
git commit -m "feat: filter the auction house to your outbid auctions"
```

---

## Task 7: The manage screen, reachable and listing

Delivers a screen the player can open from the menu, showing both lists. Inspect, cancel and refresh arrive in Tasks 8 and 9.

**Files:**
- Create: `farental-tui/src/screen/auctionhouse/manage/manage.go`
- Modify: `farental-tui/src/internal/keybind/context.go:52` and `:700`
- Modify: `farental-tui/src/screen/auctionhouse/menu/menu.go:98-100,116-118`
- Modify: `farental-tui/src/main.go:182`

**Interfaces:**
- Consumes: `request.AuctionGetOwn()`, `request.AuctionGetBids()` — both already exist and are unused; `helper.Fetch[[]api.AuctionResponse]` returns `*[]api.AuctionResponse`.
- Produces: `manage.New() *manage.Screen` implementing `orvyn.Screen`; `keybind.ContextAuctionManage`.

- [ ] **Step 1: Add the help context**

In `farental-tui/src/internal/keybind/context.go`, beside `ContextAuctionBuy` (line 52):

```go
	ContextAuctionManage                       bubblehelp.KeymapContext = "auctionManage"
```

Match the column alignment of the surrounding constants — `gofmt` will fix it if not.

Then, next to where `auctionBuyKeymap` is registered (around line 677-700):

```go
	auctionManageKeymap := bubblehelp.NewKeymap(3)
	auctionManageKeymap.Style = mainHelpStyle
	auctionManageKeymap.NewKeyBinding(Up, false)
	auctionManageKeymap.NewKeyBinding(Down, false)
	auctionManageKeymap.NewKeyBinding(Tab, true)
	auctionManageKeymap.SetHelpDesc(Tab, lokyn.L("listings / bids"))
	auctionManageKeymap.NewKeyBinding(IKey, true)
	auctionManageKeymap.SetHelpDesc(IKey, lokyn.L("information"))
	auctionManageKeymap.NewKeyBinding(CKey, true)
	auctionManageKeymap.SetHelpDesc(CKey, lokyn.L("cancel auction"))
	auctionManageKeymap.NewKeyBinding(RKey, true)
	auctionManageKeymap.SetHelpDesc(RKey, lokyn.L("refresh"))
	auctionManageKeymap.NewKeyBinding(Esc, true)
	auctionManageKeymap.NewKeyBinding(Quit, true)
	auctionManageKeymap.NewKeyBinding(Help, true)

	bubblehelp.RegisterContext(ContextAuctionManage, auctionManageKeymap)
```

`CKey` already exists (`internal/keybind/keybind.go:36`, bound at `:143`) and is used by the dashboard and script explorer contexts — nothing new to declare.

- [ ] **Step 2: Write the screen**

Create `farental-tui/src/screen/auctionhouse/manage/manage.go`:

```go
package manage

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/context"
	"farental/internal/helper"
	"farental/internal/keybind"
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
	"github.com/halsten-dev/orvyn/widget/label"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

// ownIndex is the own-listings list's position in the screen's focus manager.
// Cancelling is offered only while it holds focus.
const ownIndex = 0

type Screen struct {
	title *orvyn.SimpleRenderable

	characterInfo *characterinfo.Widget

	lblOwn  *label.Widget
	ownList *widgetlist.Widget[api.AuctionResponse]

	lblBids  *label.Widget
	bidsList *widgetlist.Widget[api.AuctionResponse]

	statusMessage *statusmessage.Widget
	help          *help.Widget

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout
}

var _ orvyn.Screen = (*Screen)(nil)

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable("manage")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.characterInfo = characterinfo.New()

	s.lblOwn = label.New("")
	s.ownList = widgetlist.New(auctionlistitem.Constructor)
	s.ownList.SetFilterable(false)
	s.ownList.SetMinSize(orvyn.NewSize(20, 6))

	s.lblBids = label.New("")
	s.bidsList = widgetlist.New(auctionlistitem.Constructor)
	s.bidsList.SetFilterable(false)
	s.bidsList.SetMinSize(orvyn.NewSize(20, 6))

	s.statusMessage = statusmessage.New()
	s.help = help.New()

	s.focusManager = orvyn.NewFocusManager()

	// Each pane is its own column: a heading naming what it holds, and the
	// list growing into whatever height is left.
	panels := []layout.FixedRatioRenderable{
		layout.NewFixedRatioRenderable(0.5,
			layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(0, 0), 1,
				s.lblOwn, s.ownList)),
		layout.NewFixedRatioRenderable(0.5,
			layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(0, 0), 1,
				s.lblBids, s.bidsList)),
	}

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(
			orvyn.NewSize(10, 4),
			2,
			s.title,
			s.characterInfo,
			layout.NewHBoxFixedRatioLayout(0, 1, 1, panels...),
			s.statusMessage,
			s.help),
	)

	return s
}

func (s *Screen) OnEnter(any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextAuctionManage)

	s.title.SetValue(lokyn.L("Auction house"))

	s.focusManager.SetWidgets([]orvyn.Focusable{s.ownList, s.bidsList})
	s.focusManager.FocusFirst()

	s.statusMessage.Reset()

	s.updateCharacterInfo()
	s.loadLists()

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
		s.statusMessage.Reset()

		switch {
		case key.Matches(k, keybind.Esc):
			return orvyn.SwitchToPreviousScreen()
		}
	}

	return s.focusManager.Update(msg)
}

// loadLists fetches both panes. They are independent, so one failing leaves the
// other populated rather than blanking the screen.
//
// Nothing here reports success: the counts live in the headings, so there is no
// status line to weigh against an error an earlier step may have set.
func (s *Screen) loadLists() {
	s.loadOwn()
	s.loadBids()

	s.updateLabels()
}

func (s *Screen) loadOwn() bool {
	res, err := helper.Fetch[[]api.AuctionResponse](request.AuctionGetOwn())

	if err != nil {
		s.statusMessage.SetError(err)
		return false
	}

	s.ownList.SetItems(*res)
	s.ownList.FocusFirst()

	return true
}

func (s *Screen) loadBids() bool {
	res, err := helper.Fetch[[]api.AuctionResponse](request.AuctionGetBids())

	if err != nil {
		s.statusMessage.SetError(err)
		return false
	}

	s.bidsList.SetItems(*res)
	s.bidsList.FocusFirst()

	return true
}

// updateLabels carries each pane's count in its heading, so an empty pane reads
// as empty rather than as a pane that failed to load.
func (s *Screen) updateLabels() {
	s.lblOwn.SetValue(fmt.Sprintf("%s (%d)",
		lokyn.L("Your auctions"), s.ownList.Length()))

	s.lblBids.SetValue(fmt.Sprintf("%s (%d)",
		lokyn.L("Your winning bids"), s.bidsList.Length()))
}

func (s *Screen) updateCharacterInfo() bool {
	characterInfo, currency, err := context.RefreshCharacterInfo(true)

	if err != nil {
		s.statusMessage.SetError(err)
		return false
	}

	unreadMail := context.RefreshHaveUnreadMail()

	data := characterinfo.ConvertCharacterInfoResponseToData(
		characterInfo, currency, unreadMail)
	s.characterInfo.UpdateData(data)

	return true
}
```

- [ ] **Step 3: Register the screen**

In `farental-tui/src/main.go`, after the `IDAuctionHouseBuy` line (182):

```go
	orvyn.RegisterScreen(screen.IDAuctionHouseManage, manage.New())
```

Add the import `"farental/screen/auctionhouse/manage"` beside the existing `buy` and `sell` imports.

- [ ] **Step 4: Wire the menu**

In `farental-tui/src/screen/auctionhouse/menu/menu.go`, add the missing case alongside `SKey` and `BKey` in `Update`:

```go
		case key.Matches(k, keybind.MKey):
			return orvyn.SwitchScreen(screen.IDAuctionHouseManage)
```

and replace the body of `btManageOnClicked`:

```go
func (s *Screen) btManageOnClicked() tea.Cmd {
	return orvyn.SwitchScreen(screen.IDAuctionHouseManage)
}
```

- [ ] **Step 5: Verify it builds clean**

```bash
cd /home/halsten/Dev/Farental/farental-tui/src
go build ./... && go vet ./... && gofmt -l . && go test ./...
```

Expected: no output from `gofmt -l`, all packages ok.

- [ ] **Step 6: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/screen/auctionhouse/manage/ src/screen/auctionhouse/menu/menu.go \
        src/internal/keybind/context.go src/main.go
git commit -m "feat: add the auction house manage screen"
```

---

## Task 8: Inspect and refresh

**Files:**
- Modify: `farental-tui/src/screen/auctionhouse/manage/manage.go`

**Interfaces:**
- Consumes: `manage.Screen` from Task 7; `auctioninspect.New(api.AuctionResponse) *auctioninspect.Screen`.
- Produces: `(*Screen).ownFocused() bool` and `(*Screen).focusedList() *widgetlist.Widget[api.AuctionResponse]`; the `update`/`Update` split and the `DialogExitMsg` switch that Task 9 extends.

- [ ] **Step 1: Add the dialog ID and the focused-pane helpers**

At the top of `manage.go`, beside `ownIndex`:

```go
const dialogIDInspect orvyn.ScreenID = "auctionManageInspect"
```

and, after `updateLabels`:

```go
func (s *Screen) ownFocused() bool {
	return s.focusManager.TabIndex() == ownIndex
}

// focusedList is the pane the keys act on. Both panes hold the same row type,
// so everything read-only can serve either without knowing which is which.
func (s *Screen) focusedList() *widgetlist.Widget[api.AuctionResponse] {
	if s.ownFocused() {
		return s.ownList
	}

	return s.bidsList
}
```

- [ ] **Step 2: Handle the keys**

Extend the key switch in `Update`:

```go
		case key.Matches(k, keybind.IKey):
			list := s.focusedList()

			if list.Length() > 0 {
				return orvyn.OpenDialog(dialogIDInspect,
					auctioninspect.New(list.GetSelectedItem()), nil)
			}

		case key.Matches(k, keybind.RKey):
			s.loadLists()

			return nil
```

and handle the dialog exit, after the key switch:

```go
	switch msg := msg.(type) {
	case orvyn.DialogExitMsg:
		if msg.DialogID == dialogIDInspect {
			// auctioninspect switches context on entry and never switches back.
			bubblehelp.SwitchToPreviousContext()

			return nil
		}
	}
```

Add the import `"farental/screen/dialog/auctioninspect"`.

- [ ] **Step 3: Re-advertise the keys the focused pane serves**

Rename the existing `Update` to `update`, and add a wrapper above it:

```go
// Update wraps the real handler so updateFocusKeybinds runs on every path. The
// dialog exits in particular restore the previous context, which resets every
// binding back to visible, and would otherwise leave the help line offering
// keys the focused pane does not serve.
func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	cmd := s.update(msg)

	s.updateFocusKeybinds()

	return cmd
}

// updateFocusKeybinds hides cancelling unless the own-listings pane holds
// focus: the bids pane lists other players' auctions, which this screen has no
// power over.
func (s *Screen) updateFocusKeybinds() {
	bubblehelp.SetKeybindVisible(keybind.CKey, s.ownFocused())
}
```

Call `s.updateFocusKeybinds()` at the end of `OnEnter` too — `SwitchContext` there has just reset every binding back to visible.

- [ ] **Step 4: Verify it builds clean**

```bash
cd /home/halsten/Dev/Farental/farental-tui/src
go build ./... && go vet ./... && gofmt -l . && go test ./...
```

Expected: no output from `gofmt -l`, all packages ok.

- [ ] **Step 5: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/screen/auctionhouse/manage/manage.go
git commit -m "feat: inspect and refresh from the manage screen"
```

---

## Task 9: Cancel a listing

**Files:**
- Modify: `farental-tui/src/screen/auctionhouse/manage/manage.go`

**Interfaces:**
- Consumes: `(*Screen).focusedList()` from Task 8; `request.AuctionCancel(api.IDBody) *resty.Request`; `popup.NewYesNo(string)` returning a dialog whose `DialogExitMsg.Param` is `uint(1)` for yes.
- Produces: nothing further.

- [ ] **Step 1: Add the dialog ID**

Beside `dialogIDInspect`:

```go
const dialogIDCancelConfirm orvyn.ScreenID = "auctionManageCancelConfirm"
```

- [ ] **Step 2: Handle the key**

In `update`'s key switch, after the `IKey` case:

```go
		case key.Matches(k, keybind.CKey):
			if s.ownFocused() && s.ownList.Length() > 0 {
				return s.openCancelConfirm()
			}
```

- [ ] **Step 3: Handle the confirmation**

In the `DialogExitMsg` switch, turn the single `if` into a switch on the dialog ID:

```go
	switch msg := msg.(type) {
	case orvyn.DialogExitMsg:
		switch msg.DialogID {
		case dialogIDInspect:
			// auctioninspect switches context on entry and never switches back.
			bubblehelp.SwitchToPreviousContext()

			return nil

		case dialogIDCancelConfirm:
			answer, ok := msg.Param.(uint)

			if ok && answer == 1 {
				s.cancelAuction()
			}

			return nil
		}
	}
```

- [ ] **Step 4: Add the confirm and the request**

After `updateCharacterInfo`:

```go
// openCancelConfirm asks before pulling a listing. Cancelling is the one
// destructive action on this screen and it cannot be undone — the auction is
// gone and the items come back by mail — so the prompt names what it will pull.
func (s *Screen) openCancelConfirm() tea.Cmd {
	auction := s.ownList.GetSelectedItem()

	return orvyn.OpenDialog(dialogIDCancelConfirm, popup.NewYesNo(
		fmt.Sprintf(lokyn.L("Cancel the auction for %s x%d ?"),
			auction.Item.Name, auction.Quantity)), nil)
}

func (s *Screen) cancelAuction() {
	auction := s.ownList.GetSelectedItem()

	_, err := helper.SendRequest(request.AuctionCancel(api.IDBody{ID: auction.ID}))

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	// Both panes are refetched rather than the cancelled row removed locally:
	// the server is the authority on what is still live, and a bid placed on
	// one of the player's other listings meanwhile would otherwise go unseen.
	infoOK := s.updateCharacterInfo()

	s.loadLists()

	if !infoOK {
		return
	}

	s.statusMessage.SetMessage(lokyn.L("Auction cancelled, check your mail"),
		statusmessage.SuccessMessage)
}
```

Add the import `"farental/screen/dialog/popup"`.

- [ ] **Step 5: Verify it builds clean**

```bash
cd /home/halsten/Dev/Farental/farental-tui/src
go build ./... && go vet ./... && gofmt -l . && go test ./...
```

Expected: no output from `gofmt -l`, all packages ok.

- [ ] **Step 6: Manual verification**

This is the only step that needs a running backend and a real terminal, and it is the only way any of the screen work is exercised — state the result honestly, including anything that did not work.

Start the server, then run the client, and check:

1. Auction house → Manage opens from both the `m` key and the button.
2. Both panes list what they should; empty panes show `(0)` rather than looking broken.
3. Tab moves between panes; the help line drops `cancel auction` on the bids pane.
4. `i` opens the inspector from either pane, and the help line is intact after closing it.
5. `c` on an own listing prompts, cancels, and both panes plus the money line refresh.
6. On the buy screen, the new **Show** selector offers `All auctions` / `My outbid auctions`, and picking the second after being outbid lists exactly those auctions.

- [ ] **Step 7: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/screen/auctionhouse/manage/manage.go
git commit -m "feat: cancel a listing from the manage screen"
```

---

## Notes for the reviewer

- Tasks 1–4 land in `farental-cli` on branch `auction-outbid-tracking`; Tasks 5–9 land in `farental-tui` on branch `auction-manage-screen`. Two branches, two pull requests.
- Task 5 cannot be exercised end to end until Tasks 1–4 are merged and deployed to whatever backend the client points at. Its own test only asserts the query string.
- Tasks 7–9 depend on neither the server work nor Tasks 5–6: they consume `/auction/own` and `/auction/bids`, which already exist.
- The bids pane is deliberately look-only. See the **Deferred** section of the spec before "improving" it.
