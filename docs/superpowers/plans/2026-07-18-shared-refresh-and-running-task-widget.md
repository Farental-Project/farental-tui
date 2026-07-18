# Shared Refresh Helpers + Running-Task Widget Rollout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deduplicate character-info/running-task fetch logic behind two shared `internal/context` functions and a reusable `internal/ticker` component, then use them to add the running-task widget to the chat, character sheet, and script editor screens (and remove the now-unnecessary task-running gate on script screens).

**Architecture:** Two new small packages (`internal/ticker`, and two new functions in `internal/context`) sit underneath all six touched screens. Existing screens (dashboard, inventory, fight, charactersheet) are refactored to call the shared functions instead of their hand-rolled fetch code — pure behavior-preserving refactors, one screen at a time, each independently buildable and runnable. Three screens (charactersheet, chat, scripteditor) then get the `runningtask` widget added on top of that foundation.

**Tech Stack:** Go 1.25, bubbletea, the `orvyn` TUI framework (`github.com/halsten-dev/orvyn`), `lokyn` for i18n. No existing test suite in this repo (`find . -name "*_test.go"` returns nothing) — verification is `go build ./...`, `go vet ./...`, `gofmt -l`, and running the app (`go run .` from `src/`) to manually exercise the changed screen. Each task below substitutes that for the usual write-test/run-test steps.

## Global Constraints

- Follow the design in `docs/superpowers/specs/2026-07-18-shared-refresh-and-running-task-widget-design.md`.
- One `ticker.Ticker` per screen — a screen needing multiple refresh cadences composes multiple calls inside one `onFire`, never runs two `Ticker`s.
- `runningtask.Widget` needs no `UpdateData()` call — it reads `context.RunningTask` directly in `Render()`.
- Every modified file must pass `gofmt -l` (no output) and `go vet ./...` before committing.
- Preserve exact existing user-visible behavior in refactor-only tasks (3, 4, 5, 6, 8) — no behavior changes, just deduplication.

---

### Task 1: `internal/ticker` package

**Files:**
- Create: `internal/ticker/ticker.go`

**Interfaces:**
- Consumes: `orvyn.TickCmd(seconds time.Duration, tag uint) tea.Cmd` and `orvyn.TickMsg{Time time.Time, Tag uint}` from `github.com/halsten-dev/orvyn` (`/home/halsten/Dev/Libs/Go/orvyn/command.go`).
- Produces: `ticker.New(interval time.Duration, onFire func()) *Ticker`, `(*Ticker).Start() tea.Cmd`, `(*Ticker).Handle(msg tea.Msg) (handled bool, cmd tea.Cmd)`, `(*Ticker).Restart() tea.Cmd`. Every later task in this plan uses these four names exactly.

- [ ] **Step 1: Write `internal/ticker/ticker.go`**

```go
package ticker

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/orvyn"
)

// Ticker wraps orvyn's tag-based TickCmd/TickMsg contract so screens don't
// each have to hand-roll a tick constant, a tag field, and a TickMsg case.
//
// One Ticker per screen: each instance owns its own tag sequence starting at
// 0, so two Tickers in the same screen could produce colliding tags at the
// same sequence position and steal each other's tick messages. A screen that
// needs more than one refresh cadence should compose multiple calls inside a
// single onFire instead of running two Tickers.
type Ticker struct {
	interval time.Duration
	tag      uint
	onFire   func()
}

// New creates a Ticker that calls onFire every interval seconds once started.
func New(interval time.Duration, onFire func()) *Ticker {
	return &Ticker{
		interval: interval,
		onFire:   onFire,
	}
}

// Start begins the tick loop. Call from the screen's OnEnter.
func (t *Ticker) Start() tea.Cmd {
	return orvyn.TickCmd(t.interval, t.tag)
}

// Handle processes msg if it is the orvyn.TickMsg this Ticker is waiting on:
// it calls onFire and returns the re-arm command. Call from the screen's
// Update, typically in a `case orvyn.TickMsg:` arm. If msg isn't the tick
// this Ticker owns, handled is false and cmd is nil.
func (t *Ticker) Handle(msg tea.Msg) (handled bool, cmd tea.Cmd) {
	tm, ok := msg.(orvyn.TickMsg)

	if !ok || tm.Tag != t.tag {
		return false, nil
	}

	t.onFire()
	t.tag++

	return true, orvyn.TickCmd(t.interval, t.tag)
}

// Restart bumps the tag (invalidating any in-flight tick still addressed to
// the old tag) and restarts the loop. Call this after a dialog closes: while
// a dialog is open, orvyn routes every message to it, so a pending tick
// command from before the dialog opened is lost and must be re-armed.
func (t *Ticker) Restart() tea.Cmd {
	t.tag++

	return orvyn.TickCmd(t.interval, t.tag)
}
```

- [ ] **Step 2: Format and build**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && gofmt -l internal/ticker/ticker.go && go build ./... && go vet ./...`
Expected: no `gofmt -l` output, both commands exit 0.

- [ ] **Step 3: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/internal/ticker/ticker.go
git commit -m "feat: add reusable Ticker wrapping orvyn's tick/tag contract"
```

---

### Task 2: `internal/context` refresh helpers

**Files:**
- Create: `internal/context/refresh.go`

**Interfaces:**
- Consumes: `helper.Fetch[T](*resty.Request) (*T, error)`, `helper.SendRequest(*resty.Request) (*resty.Response, error)` (`internal/helper/request.go`); `request.CharacterGetInfo()`, `request.CharacterGetCurrencyAmount(api.CurrencyCode)`, `request.TaskGetRunning()` (`core/request/character.go`, `core/request/task.go`); package-level vars `CharacterID`, `CharacterInfo`, `RunningTask`, `ChatContent` (`internal/context/context.go:18-28`).
- Produces: `context.RefreshCharacterInfo(fresh bool) (*api.CharacterInfoResponse, int, error)`, `context.RefreshRunningTask() error`. Tasks 3-10 call these by exact name.

- [ ] **Step 1: Write `internal/context/refresh.go`**

```go
package context

import (
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
)

// RefreshCharacterInfo fetches the current currency amount and, when fresh is
// true or nothing is cached yet, character info from the server. It updates
// CharacterID and CharacterInfo, and — on a fresh fetch that crossed into a
// new location — clears ChatContent, mirroring the dashboard's pre-existing
// behavior. The caller feeds the returned values into a characterinfo.Widget
// via its own UpdateData call.
func RefreshCharacterInfo(fresh bool) (*api.CharacterInfoResponse, int, error) {
	info := CharacterInfo

	if fresh || info == nil {
		fetched, err := helper.Fetch[api.CharacterInfoResponse](request.CharacterGetInfo())

		if err != nil {
			return nil, 0, err
		}

		if info == nil || info.Location.ID != fetched.Location.ID {
			ChatContent = make([]string, 0)
		}

		CharacterID = fetched.ID
		CharacterInfo = fetched
		info = fetched
	}

	currencyResp, err := helper.Fetch[api.CurrencyResponse](
		request.CharacterGetCurrencyAmount(api.Grynars))

	if err != nil {
		return nil, 0, err
	}

	return info, currencyResp.Amount, nil
}

// RefreshRunningTask fetches the player's current running task, if any, and
// updates RunningTask. No widget update call is needed afterwards —
// runningtask.Widget reads RunningTask directly in its Render().
func RefreshRunningTask() error {
	resp, err := helper.SendRequest(request.TaskGetRunning())

	if err != nil {
		return err
	}

	if resp.StatusCode() == 404 {
		RunningTask = nil
		return nil
	}

	RunningTask = resp.Result().(*api.TaskResponse)

	return nil
}
```

- [ ] **Step 2: Format and build**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && gofmt -l internal/context/refresh.go && go build ./... && go vet ./...`
Expected: no `gofmt -l` output, both commands exit 0.

- [ ] **Step 3: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/internal/context/refresh.go
git commit -m "feat: add shared RefreshCharacterInfo/RefreshRunningTask helpers"
```

---

### Task 3: Migrate dashboard to the shared helpers and Ticker

**Files:**
- Modify: `screen/dashboard/data.go:31-64` (`updateData`), `screen/dashboard/data.go:171-187` (delete `updateRunningTask`)
- Modify: `screen/dashboard/dashboard.go:1-25` (imports), `:27-29` (delete `tick` const), `:32` (struct field), `:154` (`OnEnter` return), `:186-196` (`DialogExitMsg` case), `:198-206` (`TickMsg` case)

**Interfaces:**
- Consumes: `context.RefreshCharacterInfo(bool) (*api.CharacterInfoResponse, int, error)`, `context.RefreshRunningTask() error` (Task 2); `ticker.New(time.Duration, func()) *ticker.Ticker`, `(*ticker.Ticker).Start/Handle/Restart` (Task 1).
- Produces: nothing new consumed elsewhere — this task is the reference implementation that tasks 7, 9, 10 mirror.

- [ ] **Step 1: Rewrite `updateData` and delete `updateRunningTask` in `data.go`**

Replace (`data.go:31-64`):

```go
func (s *Screen) updateData() {
	var err error

	defer s.updateErr(&err)

	characterInfo, err := helper.Fetch[api.CharacterInfoResponse](request.CharacterGetInfo())

	if err != nil {
		return
	}

	// If the character changed of location
	if context.CharacterInfo == nil ||
		context.CharacterInfo.Location.ID != characterInfo.Location.ID {
		context.ChatContent = make([]string, 0)
	}

	context.CharacterID = characterInfo.ID
	context.CharacterInfo = characterInfo

	currencyResp, err := helper.Fetch[api.CurrencyResponse](
		request.CharacterGetCurrencyAmount(api.Grynars))

	if err != nil {
		return
	}

	s.characterInfo.UpdateData(characterInfo, currencyResp.Amount)
	s.locationInfo.UpdateData(&characterInfo.Location)
	s.updateRunningTask()
	s.updateEventLog()
	s.updateChat()
	s.updateVisibleCharacters()
}
```

with:

```go
func (s *Screen) updateData() {
	var err error

	defer s.updateErr(&err)

	characterInfo, currency, err := context.RefreshCharacterInfo(true)

	if err != nil {
		return
	}

	s.characterInfo.UpdateData(characterInfo, currency)
	s.locationInfo.UpdateData(&characterInfo.Location)

	if refreshErr := context.RefreshRunningTask(); refreshErr != nil {
		log.Println(refreshErr)
	}

	s.updateEventLog()
	s.updateChat()
	s.updateVisibleCharacters()
}
```

Then delete the now-unused `updateRunningTask` function (`data.go:171-187`):

```go
func (s *Screen) updateRunningTask() {
	resp, err := helper.SendRequest(request.TaskGetRunning())

	if err != nil {
		log.Println(err)
		return
	}

	if resp.StatusCode() == 404 {
		context.RunningTask = nil
		return
	}

	task := resp.Result().(*api.TaskResponse)

	context.RunningTask = task
}
```

`data.go`'s imports are unchanged — `api`, `request`, `helper`, and `log` are all still used by `updateEventLog`/`updateVisibleCharacters` and the new `log.Println` call.

- [ ] **Step 2: Migrate `dashboard.go` to `ticker.Ticker`**

Add to the import block (`dashboard.go:1-25`), alongside the other `farental/...` imports:

```go
	"farental/internal/ticker"
```

Delete the const block (`dashboard.go:27-29`):

```go
const (
	tick time.Duration = 60
)
```

Change the struct field (`dashboard.go:32`) from:

```go
	tickTag uint
```

to:

```go
	ticker *ticker.Ticker
```

In `New()`, initialize it (add near the other widget/field initialization, before `return s`):

```go
	s.ticker = ticker.New(60, s.updateData)
```

Change `OnEnter`'s return (`dashboard.go:154`) from:

```go
	return tea.Batch(cmd, orvyn.TickCmd(tick, s.tickTag))
```

to:

```go
	return tea.Batch(cmd, s.ticker.Start())
```

Change the `DialogExitMsg` case (`dashboard.go:186-196`) from:

```go
	case orvyn.DialogExitMsg:
		if msg.DialogID == "earlyClaimConfirm" {
			if msg.Param.(uint) == 1 {
				s.doClaim()
			}
		}

		// While a dialog is open orvyn routes every message to it, so the
		// dashboard's spinner and refresh tick loops stop being fed and die.
		// Re-arm both on close, mirroring OnEnter, or the animation freezes.
		return tea.Batch(s.runningTask.Init(), orvyn.TickCmd(tick, s.tickTag))
```

to:

```go
	case orvyn.DialogExitMsg:
		if msg.DialogID == "earlyClaimConfirm" {
			if msg.Param.(uint) == 1 {
				s.doClaim()
			}
		}

		// While a dialog is open orvyn routes every message to it, so the
		// dashboard's spinner and refresh tick loops stop being fed and die.
		// Re-arm both on close, mirroring OnEnter, or the animation freezes.
		return tea.Batch(s.runningTask.Init(), s.ticker.Restart())
```

Change the `TickMsg` case (`dashboard.go:198-206`) from:

```go
	case orvyn.TickMsg:
		if msg.Tag != s.tickTag {
			return nil
		}

		s.updateData()

		s.tickTag++
		return orvyn.TickCmd(tick, s.tickTag)
```

to:

```go
	case orvyn.TickMsg:
		handled, cmd := s.ticker.Handle(msg)

		if !handled {
			return nil
		}

		return cmd
```

Note: `dashboard.go` still needs the `time` import for the unrelated `lastEventLogTimestamp time.Time` field — do not remove it.

- [ ] **Step 3: Format and build**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && gofmt -l screen/dashboard/data.go screen/dashboard/dashboard.go && go build ./... && go vet ./...`
Expected: no `gofmt -l` output, both commands exit 0.

- [ ] **Step 4: Manually verify no regression**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go run .`
Log in, reach the dashboard, and confirm: character info/HP/MP bars show, running-task spinner/remaining-time still works if a task is running, claiming an unfinished task still shows the confirm dialog and works for both Yes/No, and the dashboard keeps refreshing (event log/chat/visible characters) without freezing after the confirm dialog closes.

- [ ] **Step 5: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/screen/dashboard/data.go src/screen/dashboard/dashboard.go
git commit -m "refactor: migrate dashboard to shared refresh helpers and Ticker"
```

---

### Task 4: Migrate inventory to `RefreshCharacterInfo`

**Files:**
- Modify: `screen/inventory/inventory.go:123-142` (`updateCharacterInfo`)

**Interfaces:**
- Consumes: `context.RefreshCharacterInfo(true)` (Task 2).

- [ ] **Step 1: Rewrite `updateCharacterInfo`**

Replace (`inventory.go:123-142`):

```go
func (s *Screen) updateCharacterInfo() {
	characterInfo, err := helper.Fetch[api.CharacterInfoResponse](request.CharacterGetInfo())

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	context.CharacterInfo = characterInfo

	currencyResp, err := helper.Fetch[api.CurrencyResponse](
		request.CharacterGetCurrencyAmount(api.Grynars))

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.characterInfo.UpdateData(context.CharacterInfo, currencyResp.Amount)
}
```

with:

```go
func (s *Screen) updateCharacterInfo() {
	characterInfo, currency, err := context.RefreshCharacterInfo(true)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.characterInfo.UpdateData(characterInfo, currency)
}
```

All of `inventory.go`'s other imports (`api`, `request`, `helper`, `context`) stay — each is still used elsewhere in the file (`loadInventory`, `loadEquippedInventory`, `useItem`, `equipItem`, `unequipItem`, `checkRunningTask`).

- [ ] **Step 2: Format and build**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && gofmt -l screen/inventory/inventory.go && go build ./... && go vet ./...`
Expected: no `gofmt -l` output, both commands exit 0.

- [ ] **Step 3: Manually verify no regression**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go run .`
Open the inventory screen, confirm HP/MP/money still display, and that using or equipping/unequipping an item still refreshes the character-info bar afterward.

- [ ] **Step 4: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/screen/inventory/inventory.go
git commit -m "refactor: migrate inventory to shared RefreshCharacterInfo"
```

---

### Task 5: Migrate fight to `RefreshCharacterInfo`

**Files:**
- Modify: `screen/fight/fight.go:62-77` (`updateData`)

**Interfaces:**
- Consumes: `context.RefreshCharacterInfo(false)` (Task 2).

- [ ] **Step 1: Rewrite `updateData`**

Replace (`fight.go:62-77`):

```go
func (s *Screen) updateData() {
	characterInfo := context.CharacterInfo

	req := request.CharacterGetCurrencyAmount(api.Grynars)

	currencyResp, err := helper.Fetch[api.CurrencyResponse](req)

	if err != nil {
		s.Screen.SetStatusError(err)
		return
	}

	s.characterInfo.UpdateData(characterInfo, currencyResp.Amount)

	s.characterActiveScript.UpdateData()
}
```

with:

```go
func (s *Screen) updateData() {
	characterInfo, currency, err := context.RefreshCharacterInfo(false)

	if err != nil {
		s.Screen.SetStatusError(err)
		return
	}

	s.characterInfo.UpdateData(characterInfo, currency)

	s.characterActiveScript.UpdateData()
}
```

`fight.go`'s `api`, `request`, and `helper` imports stay — all three are still used by `loadFights` and `submit` (`fight.go:105,109,134,136`).

- [ ] **Step 2: Format and build**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && gofmt -l screen/fight/fight.go && go build ./... && go vet ./...`
Expected: no `gofmt -l` output, both commands exit 0.

- [ ] **Step 3: Manually verify no regression**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go run .`
Open the fight list screen and confirm the character-info header still shows HP/MP/money and the active-script summary.

- [ ] **Step 4: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/screen/fight/fight.go
git commit -m "refactor: migrate fight screen to shared RefreshCharacterInfo"
```

---

### Task 6: Migrate charactersheet to `RefreshCharacterInfo`

**Files:**
- Modify: `screen/charactersheet/charactersheet.go:1-26` (imports), `:134-155` (`updateData`)

**Interfaces:**
- Consumes: `context.RefreshCharacterInfo(false)` (Task 2).

- [ ] **Step 1: Rewrite `updateData`**

Replace (`charactersheet.go:134-155`):

```go
func (s *Screen) updateData() {
	characterInfo := context.CharacterInfo

	req := request.CharacterGetCurrencyAmount(api.Grynars)

	currencyResp, err := helper.Fetch[api.CurrencyResponse](req)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.characterInfo.UpdateData(characterInfo, currencyResp.Amount)

	s.characterActiveScript.UpdateData()

	s.equipmentSummary.UpdateData()

	s.statsSummary.UpdateData(characterInfo)

	s.skillsSummary.UpdateData(characterInfo)
}
```

with:

```go
func (s *Screen) updateData() {
	characterInfo, currency, err := context.RefreshCharacterInfo(false)

	if err != nil {
		s.statusMessage.SetError(err)
		return
	}

	s.characterInfo.UpdateData(characterInfo, currency)

	s.characterActiveScript.UpdateData()

	s.equipmentSummary.UpdateData()

	s.statsSummary.UpdateData(characterInfo)

	s.skillsSummary.UpdateData(characterInfo)
}
```

- [ ] **Step 2: Remove now-unused imports**

`api`, `request`, and `helper` were only used inside the function just rewritten (confirmed via `grep -n "api\.\|request\.\|helper\." screen/charactersheet/charactersheet.go`, which only matched inside the old `updateData`). Remove these three lines from the import block (`charactersheet.go:4,5,7`):

```go
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/helper"
```

`context` stays (still used by `context.RefreshCharacterInfo`).

- [ ] **Step 3: Format and build**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && gofmt -w screen/charactersheet/charactersheet.go && go build ./... && go vet ./...`
Expected: both commands exit 0. (Using `gofmt -w` here, not `-l`, to auto-fix import block ordering after removing lines.)

- [ ] **Step 4: Manually verify no regression**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go run .`
Open the character sheet screen and confirm HP/MP/money, active script, equipment summary, and stats/skills all still display.

- [ ] **Step 5: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/screen/charactersheet/charactersheet.go
git commit -m "refactor: migrate charactersheet to shared RefreshCharacterInfo"
```

---

### Task 7: Add running-task widget to charactersheet

**Files:**
- Modify: `screen/charactersheet/charactersheet.go` (imports, struct, `New()`, `OnEnter`, `Update`)

**Interfaces:**
- Consumes: `runningtask.New() *runningtask.Widget`, `(*runningtask.Widget).Init() tea.Cmd`, `(*runningtask.Widget).Update(tea.Msg) tea.Cmd` (`widget/runningtask/runningtask.go`); `ticker.New`/`Start`/`Handle`/`Restart` (Task 1); `context.RefreshRunningTask()` (Task 2).

- [ ] **Step 1: Add imports**

Add to the import block (alphabetical position among the `farental/...` group, plus stdlib `log`):

```go
	"farental/internal/ticker"
	...
	"farental/widget/runningtask"
	...
	"log"
```

(Run `gofmt -w` at the end of this task to fix exact ordering — don't hand-sort.)

- [ ] **Step 2: Add struct fields**

In `type Screen struct { ... }`, add two fields (near `characterInfo *characterinfo.Widget`):

```go
	runningTask *runningtask.Widget

	ticker *ticker.Ticker
```

- [ ] **Step 3: Initialize the widget and ticker in `New()`**

Add, right after `s.characterInfo = characterinfo.New()`:

```go
	s.runningTask = runningtask.New()
```

Add, before `return s`:

```go
	s.ticker = ticker.New(60, func() {
		if err := context.RefreshRunningTask(); err != nil {
			log.Println(err)
		}
	})
```

- [ ] **Step 4: Insert the widget into the layout**

Change the layout's child list from:

```go
			s.title,
			orvyn.VGap,
			s.characterInfo,
			s.characterActiveScript,
			s.equipmentSummary,
			s.statsSkillLayout,
			s.statusMessage,
			s.help,
```

to:

```go
			s.title,
			orvyn.VGap,
			s.runningTask,
			s.characterInfo,
			s.characterActiveScript,
			s.equipmentSummary,
			s.statsSkillLayout,
			s.statusMessage,
			s.help,
```

- [ ] **Step 5: Start the ticker and widget in `OnEnter`**

Change `OnEnter`'s body from:

```go
func (s *Screen) OnEnter(i any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextCharacterSheet)

	s.title.SetValue(lokyn.L("Character"))

	s.statusMessage.Reset()

	orvyn.SetPreviousScreen(screen.IDDashBoard)

	s.updateData()

	return nil
}
```

to:

```go
func (s *Screen) OnEnter(i any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextCharacterSheet)

	s.title.SetValue(lokyn.L("Character"))

	s.statusMessage.Reset()

	orvyn.SetPreviousScreen(screen.IDDashBoard)

	s.updateData()

	if err := context.RefreshRunningTask(); err != nil {
		log.Println(err)
	}

	return tea.Batch(s.runningTask.Init(), s.ticker.Start())
}
```

- [ ] **Step 6: Handle ticks and drive the spinner in `Update`**

Change `Update` from:

```go
func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return tea.Quit

		case key.Matches(msg, keybind.Esc):
			return orvyn.SwitchToPreviousScreen()

		case key.Matches(msg, keybind.IKey):
			return orvyn.SwitchScreen(screen.IDInventory)

		}
	}

	s.skillsSummary.Update(msg)

	return nil
}
```

to:

```go
func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Quit):
			return tea.Quit

		case key.Matches(msg, keybind.Esc):
			return orvyn.SwitchToPreviousScreen()

		case key.Matches(msg, keybind.IKey):
			return orvyn.SwitchScreen(screen.IDInventory)

		}

	case orvyn.TickMsg:
		handled, cmd := s.ticker.Handle(msg)

		if !handled {
			return nil
		}

		return cmd
	}

	s.skillsSummary.Update(msg)

	return s.runningTask.Update(msg)
}
```

- [ ] **Step 7: Format and build**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && gofmt -w screen/charactersheet/charactersheet.go && go build ./... && go vet ./...`
Expected: both commands exit 0.

- [ ] **Step 8: Manually verify**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go run .`
Start a task from the dashboard (e.g. travel), then open the character sheet screen: confirm the running-task widget shows the task title, spinner, and remaining time, and that the remaining time visibly ticks down the longer you stay on the screen (wait past 60s, or reduce the interval locally to `2` for a quick manual check and revert before committing).

- [ ] **Step 9: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/screen/charactersheet/charactersheet.go
git commit -m "feat: show running-task widget on the character sheet screen"
```

---

### Task 8: Migrate chat's tick to `ticker.Ticker`

**Files:**
- Modify: `screen/chat/chat.go:1-28` (imports, const), `:31` (struct field), `:92-102` (`OnEnter`), `:126-134` (`TickMsg` case)

**Interfaces:**
- Consumes: `ticker.New`/`Start`/`Handle` (Task 1).

- [ ] **Step 1: Replace the `time` import with `ticker`**

`time` (`chat.go:13`) is only used by the `tick time.Duration = 15` const being removed in this task (confirmed via `grep -n "time\." screen/chat/chat.go`, which only matches that const line). Remove `"time"` from the import block and add `"farental/internal/ticker"` (`gofmt -w` will fix ordering at the end of this task).

- [ ] **Step 2: Delete the tick const**

Delete (`chat.go:26-28`):

```go
const (
	tick time.Duration = 15
)
```

- [ ] **Step 3: Replace the `tickTag` field**

Change (`chat.go:31`):

```go
	tickTag uint
```

to:

```go
	ticker *ticker.Ticker
```

- [ ] **Step 4: Initialize the ticker in `New()`**

Add, before `return s`:

```go
	s.ticker = ticker.New(15, s.loadChat)
```

- [ ] **Step 5: Update `OnEnter`**

Change (`chat.go:92-102`):

```go
func (s *Screen) OnEnter(i any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextChat)

	s.title.SetValue(lokyn.L("Chat"))

	cmd := s.input.Init()

	s.loadChat()

	return tea.Batch(orvyn.TickCmd(tick, s.tickTag), cmd)
}
```

to:

```go
func (s *Screen) OnEnter(i any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextChat)

	s.title.SetValue(lokyn.L("Chat"))

	cmd := s.input.Init()

	s.loadChat()

	return tea.Batch(s.ticker.Start(), cmd)
}
```

- [ ] **Step 6: Update the `TickMsg` case**

Change (`chat.go:126-134`):

```go
	case orvyn.TickMsg:
		if msg.Tag != s.tickTag {
			return nil
		}

		s.loadChat()

		s.tickTag++
		return orvyn.TickCmd(tick, s.tickTag)

```

to:

```go
	case orvyn.TickMsg:
		handled, cmd := s.ticker.Handle(msg)

		if !handled {
			return nil
		}

		return cmd

```

- [ ] **Step 7: Format and build**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && gofmt -w screen/chat/chat.go && go build ./... && go vet ./...`
Expected: both commands exit 0.

- [ ] **Step 8: Manually verify no regression**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go run .`
Open the chat screen and confirm messages still load on entry and the log still auto-refreshes while you sit on the screen (send a message from another client/session if available, or just confirm no crash/freeze over ~20s).

- [ ] **Step 9: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/screen/chat/chat.go
git commit -m "refactor: migrate chat screen's tick to shared Ticker"
```

---

### Task 9: Add running-task widget to chat

**Files:**
- Modify: `screen/chat/chat.go` (imports, struct, `New()`, layout, `OnEnter`, `Update`)

**Interfaces:**
- Consumes: `runningtask.New/Init/Update` (`widget/runningtask`); `context.RefreshRunningTask()` (Task 2).

- [ ] **Step 1: Add imports**

Add `"farental/widget/runningtask"` and `"log"` to the import block (`gofmt -w` fixes ordering at the end).

- [ ] **Step 2: Add the struct field**

Add to `type Screen struct { ... }`, near `logChat *simplelogviewer.Widget`:

```go
	runningTask *runningtask.Widget
```

- [ ] **Step 3: Initialize it in `New()`**

Add, near `s.logChat = simplelogviewer.New("")`:

```go
	s.runningTask = runningtask.New()
```

- [ ] **Step 4: Insert it into the layout**

Change the layout's child list from:

```go
			s.title,
			orvyn.VGap,
			s.logChat,
			s.input,
			s.statusMessage,
			s.help,
```

to:

```go
			s.title,
			orvyn.VGap,
			s.runningTask,
			s.logChat,
			s.input,
			s.statusMessage,
			s.help,
```

- [ ] **Step 5: Fold the running-task refresh into the ticker's `onFire`**

Change the ticker initialization in `New()` from:

```go
	s.ticker = ticker.New(15, s.loadChat)
```

to:

```go
	s.ticker = ticker.New(15, s.refreshData)
```

Add a new method (near `loadChat`):

```go
func (s *Screen) refreshData() {
	s.loadChat()

	if err := context.RefreshRunningTask(); err != nil {
		log.Println(err)
	}
}
```

- [ ] **Step 6: Start the widget in `OnEnter`**

Change `OnEnter` from:

```go
func (s *Screen) OnEnter(i any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextChat)

	s.title.SetValue(lokyn.L("Chat"))

	cmd := s.input.Init()

	s.loadChat()

	return tea.Batch(s.ticker.Start(), cmd)
}
```

to:

```go
func (s *Screen) OnEnter(i any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextChat)

	s.title.SetValue(lokyn.L("Chat"))

	cmd := s.input.Init()

	s.loadChat()

	if err := context.RefreshRunningTask(); err != nil {
		log.Println(err)
	}

	return tea.Batch(s.ticker.Start(), s.runningTask.Init(), cmd)
}
```

- [ ] **Step 7: Drive the spinner in `Update`**

Change the final lines of `Update` from:

```go
	cmd := s.input.Update(msg)

	s.logChat.Update(msg)

	return cmd
}
```

to:

```go
	cmd := s.input.Update(msg)

	s.logChat.Update(msg)

	return tea.Batch(cmd, s.runningTask.Update(msg))
}
```

- [ ] **Step 8: Format and build**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && gofmt -w screen/chat/chat.go && go build ./... && go vet ./...`
Expected: both commands exit 0.

- [ ] **Step 9: Manually verify**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go run .`
Start a task from the dashboard, open the chat screen, confirm the running-task widget shows and its remaining time ticks down over time (same 15s cadence as chat refresh).

- [ ] **Step 10: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/screen/chat/chat.go
git commit -m "feat: show running-task widget on the chat screen"
```

---

### Task 10: Add running-task widget to scripteditor

**Files:**
- Modify: `screen/scripteditor/scripteditor.go` (imports, struct, `New()`, layout, `OnEnter`, `Update`)

**Interfaces:**
- Consumes: `runningtask.New/Init/Update`; `ticker.New/Start/Handle/Restart`; `context.RefreshRunningTask()`.

- [ ] **Step 1: Add imports**

Add `"farental/internal/context"`, `"farental/internal/ticker"`, and `"farental/widget/runningtask"` to the import block (`gofmt -w` fixes ordering at the end).

- [ ] **Step 2: Add struct fields**

Add to `type Screen struct { ... }`, near `statusMessage *statusmessage.Widget`:

```go
	runningTask *runningtask.Widget

	ticker *ticker.Ticker
```

- [ ] **Step 3: Initialize in `New()`**

Add, near `s.statusMessage = statusmessage.New()`:

```go
	s.runningTask = runningtask.New()
```

Add, before `return s`:

```go
	s.ticker = ticker.New(60, func() {
		if err := context.RefreshRunningTask(); err != nil {
			log.Println(err)
		}
	})
```

(`"log"` is already imported by `scripteditor.go`.)

- [ ] **Step 4: Insert it into the layout**

Change the layout's child list from:

```go
			s.title,
			s.readOnlyTitle,
			orvyn.VGap,
			layout.NewHBoxFixedRatioLayout(0, 1, 1, inspectorElements...),
			s.statusMessage,
			s.help,
```

to:

```go
			s.title,
			s.readOnlyTitle,
			orvyn.VGap,
			s.runningTask,
			layout.NewHBoxFixedRatioLayout(0, 1, 1, inspectorElements...),
			s.statusMessage,
			s.help,
```

- [ ] **Step 5: Start the widget/ticker in `OnEnter`**

`OnEnter` currently returns `nil` at the very end (`scripteditor.go:159`) — it also has an earlier `return orvyn.SwitchToPreviousScreen()` on script-fetch failure (`scripteditor.go:132`), which should NOT start the ticker since the screen is being exited. Change only the final line, from:

```go
	s.focusManager.FocusFirst()

	return nil
}
```

to:

```go
	s.focusManager.FocusFirst()

	if err := context.RefreshRunningTask(); err != nil {
		log.Println(err)
	}

	return tea.Batch(s.runningTask.Init(), s.ticker.Start())
}
```

- [ ] **Step 6: Handle ticks and re-arm on dialog close in `Update`**

Change the `orvyn.DialogExitMsg` case (`scripteditor.go:214-225`) from:

```go
	case orvyn.DialogExitMsg:
		switch msg.DialogID {
		case "quitConfirm":
			val := msg.Param.(uint)

			switch val {
			case 1:
				return orvyn.SwitchScreen(screen.IDScriptExplorer)
			default:
				return nil
			}
		}

	}
```

to:

```go
	case orvyn.DialogExitMsg:
		switch msg.DialogID {
		case "quitConfirm":
			val := msg.Param.(uint)

			switch val {
			case 1:
				return orvyn.SwitchScreen(screen.IDScriptExplorer)
			default:
				return s.ticker.Restart()
			}
		}

	case orvyn.TickMsg:
		handled, cmd := s.ticker.Handle(msg)

		if !handled {
			return nil
		}

		return cmd

	}
```

(The `quitConfirm` dialog only fires when leaving the editor without saving; choosing "Yes" switches screens away, so only the "No" / default branch needs to re-arm the ticker.)

- [ ] **Step 7: Drive the spinner**

Change `Update`'s final lines from:

```go
	cmd := s.focusManager.Update(msg)

	return cmd
}
```

to:

```go
	cmd := s.focusManager.Update(msg)

	return tea.Batch(cmd, s.runningTask.Update(msg))
}
```

- [ ] **Step 8: Format and build**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && gofmt -w screen/scripteditor/scripteditor.go && go build ./... && go vet ./...`
Expected: both commands exit 0.

- [ ] **Step 9: Manually verify**

Deferred to Task 11's manual check, since the script editor isn't reachable with a running task until that task removes the gate. `go build`/`go vet` passing is sufficient here.

- [ ] **Step 10: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/screen/scripteditor/scripteditor.go
git commit -m "feat: show running-task widget on the script editor screen"
```

---

### Task 11: Remove the running-task gate on script screens

**Files:**
- Modify: `screen/dashboard/dashboard.go:315-318` (`gameKeyHandler`, `SKey` case)

**Interfaces:** none new.

- [ ] **Step 1: Remove the gate**

Change:

```go
	case key.Matches(msg, keybind.SKey):
		if s.checkRunningTask() {
			return orvyn.SwitchScreen(screen.IDScriptExplorer), true
		}
```

to:

```go
	case key.Matches(msg, keybind.SKey):
		return orvyn.SwitchScreen(screen.IDScriptExplorer), true
```

`checkRunningTask()` stays defined and used by the other gated actions (`LKey`, `TKey`, `AKey`, `FKey`, `CKey`, `NKey`) — do not delete it.

- [ ] **Step 2: Format and build**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && gofmt -l screen/dashboard/dashboard.go && go build ./... && go vet ./...`
Expected: no `gofmt -l` output, both commands exit 0.

- [ ] **Step 3: Manually verify end-to-end**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go run .`
Start a task from the dashboard (e.g. travel or fight), then press the script-list key while it's running: confirm you can now reach the script explorer and open a script in the editor, and the running-task widget in the editor shows the live countdown. Also spot-check that the other task-gated actions (travel/fight/craft/npc/location-services) still correctly refuse to open while a task is running (unchanged `checkRunningTask()` behavior).

- [ ] **Step 4: Commit**

```bash
cd /home/halsten/Dev/Farental/farental-tui
git add src/screen/dashboard/dashboard.go
git commit -m "feat: allow browsing/editing scripts while a task is running

Server-side confirmed no /script/* route has the NoRunningTaskRequired
middleware applied (farental-cli serverapi/routes.go), so this TUI-only
gate wasn't a real enforcement boundary."
```

---

### Task 12: Full regression pass

**Files:** none (verification only).

- [ ] **Step 1: Build and vet the whole module**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && gofmt -l . && go build ./... && go vet ./...`
Expected: no `gofmt -l` output, both commands exit 0.

- [ ] **Step 2: Walk every touched screen**

Run: `cd /home/halsten/Dev/Farental/farental-tui/src && go run .`
Confirm, in one session: login → character selection → dashboard (character info, running-task widget, claim-early confirm dialog, event log/chat/visible-characters keep refreshing) → inventory (character info updates on use/equip/unequip) → fight list (character info + active script header) → character sheet (character info + running-task widget with live countdown) → chat (character info independent of this change, running-task widget with live countdown, messages still load) → script explorer/editor reachable and functional with a task running, running-task widget visible there too, and the quit-confirm dialog in the editor still works and doesn't freeze the countdown afterward.

- [ ] **Step 3: Report**

No commit for this task — it's a checkpoint. If anything regressed, fix it in the relevant task's screen and amend that task's commit-worthy fix as a new small commit, then re-run this task.
