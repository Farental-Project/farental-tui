# Shared character-info/running-task refresh helpers + running-task widget rollout

Date: 2026-07-18

## Problem

The `characterinfo` widget is used in four screens today (dashboard, inventory,
fight, charactersheet), each with its own hand-rolled fetch code that has
drifted into three different behaviors:

- **dashboard** (`screen/dashboard/data.go`): fetches character info fresh
  from the server plus currency, on a 60s tick and on enter.
- **inventory** (`screen/inventory/inventory.go`): fetches fresh character
  info + currency, but only on enter and after inventory-mutating actions
  (equip/drop), no tick.
- **fight** / **charactersheet**: reuse the cached `context.CharacterInfo`
  (no character-info network call), only refetch currency, on enter only.

Separately, the `runningtask` widget (spinner + remaining time) is only used
in dashboard. The widget itself is stateless about data — it reads
`context.RunningTask` directly in `Render()` — but nothing refreshes that
context value except dashboard's own 60s tick.

We want to (1) add the `runningtask` widget to the chat, character sheet,
and script editor screens, and (2) deduplicate the character-info fetch
logic that's currently copy-pasted with three different behaviors across four
screens.

## Design

### `internal/context/refresh.go` (new file)

Two new functions, living in `internal/context` (the same package that
already hosts `UpdateChat()`, a fetch-and-mutate-global-state function of the
same shape):

```go
// RefreshCharacterInfo fetches currency, and character info if fresh is true
// or nothing is cached yet. Returns the info and currency amount for the
// caller to feed into a characterinfo.Widget.
func RefreshCharacterInfo(fresh bool) (*api.CharacterInfoResponse, int, error)

// RefreshRunningTask fetches the current running task (if any) and updates
// context.RunningTask. No widget update needed — runningtask.Widget reads
// context.RunningTask directly.
func RefreshRunningTask() error
```

`RefreshCharacterInfo` preserves dashboard's existing side effect of clearing
`context.ChatContent` when the character's location changes, scoped to the
`fresh` path (that comparison is meaningless when reusing the cache).

This replaces the fetch logic in:
- `screen/dashboard/data.go` (`updateData`, `updateRunningTask`)
- `screen/inventory/inventory.go` (`updateCharacterInfo`)
- `screen/fight/fight.go` (`updateData`)
- `screen/charactersheet/charactersheet.go` (`updateData`)

Each call site keeps its own `widget.UpdateData(info, currency)` call — the
widget update stays in the screen layer, only the fetch+state-mutation is
shared.

### `internal/ticker` (new package)

A small wrapper around `orvyn.TickCmd`/`orvyn.TickMsg`'s tag-based contract
(see `github.com/halsten-dev/orvyn/command.go`), which requires the tag to be
bumped on every re-arm to invalidate stale in-flight ticks:

```go
type Ticker struct { /* interval, tag, onFire */ }

func New(interval time.Duration, onFire func()) *Ticker
func (t *Ticker) Start() tea.Cmd                        // call from OnEnter
func (t *Ticker) Handle(msg tea.Msg) (handled bool, cmd tea.Cmd) // call from Update
func (t *Ticker) Restart() tea.Cmd                       // call after a DialogExitMsg
```

**Constraint: one `Ticker` per screen.** Each ticker owns its own tag
sequence starting at 0; two tickers in the same screen could produce
colliding tags at the same sequence position and steal each other's tick
messages. A screen that needs multiple refresh cadences composes multiple
calls inside one `onFire`, rather than running two tickers.

### Per-screen changes

- **dashboard**: `data.go` calls `context.RefreshCharacterInfo(true)` and
  `context.RefreshRunningTask()` instead of hand-rolled fetch code. Its
  existing tick (`dashboard.go` `tick`/`tickTag`) migrates to `ticker.Ticker`,
  collapsing the `DialogExitMsg` re-arm (`dashboard.go:186-196`) to a single
  `ticker.Restart()` call.
- **inventory**: `updateCharacterInfo()` becomes
  `context.RefreshCharacterInfo(true)` + `s.characterInfo.UpdateData(...)`.
- **fight**, **charactersheet**: `updateData()`'s character-info/currency
  block becomes `context.RefreshCharacterInfo(false)` +
  `s.characterInfo.UpdateData(...)`.
- **charactersheet**, **chat**: add a `runningtask.Widget` field, a layout
  slot for it, and a `ticker.Ticker` at 60s whose `onFire` calls
  `context.RefreshRunningTask()`. Chat's existing hand-rolled 15s tick
  (`chat.go` `tick`/`tickTag`) migrates to `ticker.Ticker`, with `loadChat()`
  and the new refresh both inside its `onFire`.
- **scripteditor**: same widget + ticker addition as charactersheet. Since it
  already has a `"quitConfirm"` dialog (`scripteditor.go:181-183`), add
  `ticker.Restart()` inside its existing `orvyn.DialogExitMsg` case
  (`scripteditor.go:214-225`), mirroring dashboard's re-arm pattern.
- **`screen/dashboard/dashboard.go` `gameKeyHandler`**: remove the
  `checkRunningTask()` gate on the `S` key (line ~315-318) so script screens
  are reachable while a task is running.

### Removing the script-screen task gate — server-side check

Confirmed against the server repo (`farental-cli`, read-only check): the
`NoRunningTaskRequired` middleware
(`serverapi/middleware/norunningtaskrequired.go`) — the server's standard
guard for "can't do this with a task running", applied to travel/craft/fight
start, inventory actions, bank, tavern, mail — is **not** applied to any
`/script/*` route (`serverapi/routes.go:68-117`). Script save/delete/setActive
are already reachable via direct API calls while a task is running today, so
the TUI-side gate is not a real enforcement boundary. Removing it doesn't
newly expose anything.

**Caveat (informational only, not a blocker):** for Fight tasks specifically,
the resolution engine (`serverapi/system/fight/engine.go`, singleton loop,
one fight resolved at a time globally via `FindUnresolvedWithPreloads`,
`repository/fight.go:13-33`) loads each fighter's script fresh via
`LoadActorsScript` (`fight/data/fight.go:110-131,140-163`) at resolve time,
not at fight start (`FightStart`, `serverapi/controller/fight.go:133-270`,
only checks a script exists, doesn't snapshot it). So editing/reactivating
the active script while a fight is queued-but-unresolved can affect that
fight's outcome. Pre-existing server behavior, unrelated to the TUI gate,
worth flagging to the server side separately.

## Error handling

Both new `context` functions return `error`. Screens surface it the same way
they do today — `s.statusMessage.SetError(err)` for user-triggered fetches.
For ticker-driven periodic `RefreshRunningTask()` calls, keep dashboard's
existing choice of a quiet `log.Println` on failure rather than spamming the
status bar every tick.

## Testing / verification

`go build ./...` and `go vet ./...`, then run the app and manually confirm:
- dashboard behaves identically to before the refactor.
- chat, character sheet, and script editor show the running-task widget with
  a live countdown.
- script screens are reachable while a fight task is running.
- scripteditor's quit-confirm dialog still correctly re-arms its tick after
  closing.

## Out of scope

- Changing the fight-engine's script-snapshot timing server-side (flagged as
  a separate follow-up, not part of this change).
- A generic multi-ticker-per-screen abstraction (YAGNI — no current screen
  needs more than one cadence once folded into a single `onFire`).
