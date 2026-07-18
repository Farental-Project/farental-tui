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
