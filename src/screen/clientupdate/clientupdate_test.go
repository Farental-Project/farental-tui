package clientupdate

import (
	"errors"
	"os"
	"strings"
	"testing"

	"farental/internal/keybind"
	"farental/internal/updater"
	"farental/screen"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
)

// dummyScreen is a no-op orvyn.Screen, registered under screen.IDLogin so that
// handleKey's orvyn.SwitchScreen(screen.IDLogin) call has somewhere real to
// go: an unregistered ScreenID makes orvyn.SwitchScreen call log.Fatalf,
// which exits the test process outright.
type dummyScreen struct{}

func (dummyScreen) OnEnter(any) tea.Cmd    { return nil }
func (dummyScreen) OnExit() any            { return nil }
func (dummyScreen) Update(tea.Msg) tea.Cmd { return nil }
func (dummyScreen) Render() orvyn.Layout   { return nil }

func TestMain(m *testing.M) {
	// lokyn.L (used throughout this package and by keybind.Init) panics on a
	// nil localizer unless the package has been initialized at least once;
	// no translation file needs to be loaded for it to fall back to the
	// English key text.
	lokyn.Init()
	lokyn.SetLanguage("en")

	keybind.Init()

	orvyn.Init()
	orvyn.RegisterScreen(screen.IDLogin, dummyScreen{})

	os.Exit(m.Run())
}

func fileInfo() updater.FileInfo {
	return updater.FileInfo{URL: "farental-linux-amd64", SHA256: "deadbeef"}
}

func TestDecideEntry(t *testing.T) {
	tests := []struct {
		name         string
		result       updater.Result
		preflightErr error
		wantState    state
		wantReason   entryReason
	}{
		{
			name:       "optional mode no file goes manual required",
			result:     updater.Result{Mode: updater.ModeOptional},
			wantState:  stateManualRequired,
			wantReason: reasonNoFile,
		},
		{
			name:       "mandatory mode no file goes manual required",
			result:     updater.Result{Mode: updater.ModeMandatory},
			wantState:  stateManualRequired,
			wantReason: reasonNoFile,
		},
		{
			name:         "optional mode preflight failure goes manual required",
			result:       updater.Result{Mode: updater.ModeOptional, File: fileInfo()},
			preflightErr: errors.New("read-only filesystem"),
			wantState:    stateManualRequired,
			wantReason:   reasonNotWritable,
		},
		{
			name:       "mandatory mode fetch error goes manual required",
			result:     updater.Result{Mode: updater.ModeMandatory, Err: errors.New("network down")},
			wantState:  stateManualRequired,
			wantReason: reasonFetchFailed,
		},
		{
			name:       "optional mode happy path goes to prompt",
			result:     updater.Result{Mode: updater.ModeOptional, File: fileInfo()},
			wantState:  statePrompt,
			wantReason: reasonNone,
		},
		{
			name:       "mandatory mode happy path goes to prompt",
			result:     updater.Result{Mode: updater.ModeMandatory, File: fileInfo()},
			wantState:  statePrompt,
			wantReason: reasonNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotState, gotReason := decideEntry(tt.result, tt.preflightErr)

			if gotState != tt.wantState {
				t.Errorf("state = %v, want %v", gotState, tt.wantState)
			}

			if gotReason != tt.wantReason {
				t.Errorf("reason = %v, want %v", gotReason, tt.wantReason)
			}
		})
	}
}

// A fetch error takes priority over a missing file, since an unreachable
// server also leaves File empty: without this ordering, decideEntry would
// misreport a network problem as a platform-support gap.
func TestDecideEntryFetchErrorTakesPriorityOverMissingFile(t *testing.T) {
	result := updater.Result{Mode: updater.ModeMandatory, Err: errors.New("timeout")}

	_, gotReason := decideEntry(result, nil)

	if gotReason != reasonFetchFailed {
		t.Errorf("reason = %v, want %v", gotReason, reasonFetchFailed)
	}
}

func escKeyMsg() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyEsc}
}

// TestHandleKeyManualRequiredEsc covers the defect from code review: a
// compatible client (ModeOptional) landing in stateManualRequired - because
// its platform has no published file, or the preflight check failed - must
// be able to reach the login screen with esc. A ModeMandatory client must
// have no such exit.
//
// orvyn.SwitchScreen performs the transition synchronously - it mutates the
// package-level current-screen state and only returns whatever tea.Cmd the
// destination screen's OnEnter hands back (nil, for our no-op dummy). So the
// switch is already visible in orvyn.GetCurrentScreenID() by the time
// handleKey returns; there is no separate command left to invoke.
//
// This only exercises handleKey's return value and the resulting global
// screen state. It does not exercise this Screen's own OnEnter or any
// rendering: those need a full bubbletea run loop and a real terminal, which
// is out of reach here.
func TestHandleKeyManualRequiredEsc(t *testing.T) {
	t.Run("optional mode reaches login", func(t *testing.T) {
		s := &Screen{state: stateManualRequired, result: updater.Result{Mode: updater.ModeOptional}}

		_, handled := s.handleKey(escKeyMsg())

		if !handled {
			t.Fatal("expected esc to be handled in stateManualRequired for ModeOptional")
		}

		if got := orvyn.GetCurrentScreenID(); got != screen.IDLogin {
			t.Errorf("current screen = %q, want %q", got, screen.IDLogin)
		}
	})

	t.Run("mandatory mode has no exit", func(t *testing.T) {
		before := orvyn.GetCurrentScreenID()

		s := &Screen{state: stateManualRequired, result: updater.Result{Mode: updater.ModeMandatory}}

		cmd, handled := s.handleKey(escKeyMsg())

		if handled {
			t.Error("esc must not be handled in stateManualRequired for ModeMandatory")
		}

		if cmd != nil {
			t.Error("expected no command for an unhandled key")
		}

		if got := orvyn.GetCurrentScreenID(); got != before {
			t.Errorf("current screen changed to %q, want unchanged %q", got, before)
		}
	})
}

// A plain download or checksum failure must keep the existing retry path:
// stateFailed, with the "press r" hint, and 'r' wired to retry.
func TestHandleFinishedOrdinaryFailure(t *testing.T) {
	s := New()
	s.result = updater.Result{Mode: updater.ModeOptional}

	cmd := s.handleFinished(finishedMsg{err: errors.New("checksum mismatch")})

	if cmd != nil {
		t.Errorf("expected no command, got %v", cmd)
	}

	if s.state != stateFailed {
		t.Errorf("state = %v, want stateFailed", s.state)
	}

	if got := s.status.Render(); !strings.Contains(got, lokyn.L("Press r to retry.")) {
		t.Errorf("status = %q, want the retry hint", got)
	}
}

// RollbackFailedError means selfupdate could not restore the pre-update
// binary either: there is nothing left at the target path, so a retry
// cannot work. handleFinished must land in stateUnrecoverable, name the
// saved old-binary path, and drop the retry hint rather than repeat it.
func TestHandleFinishedRollbackFailure(t *testing.T) {
	s := New()
	s.result = updater.Result{Mode: updater.ModeOptional}

	rerr := &updater.RollbackFailedError{
		OldPath: "/opt/farental/.Farental.old",
		Err:     errors.New("rename failed"),
	}

	cmd := s.handleFinished(finishedMsg{err: rerr})

	if cmd != nil {
		t.Errorf("expected no command, got %v", cmd)
	}

	if s.state != stateUnrecoverable {
		t.Errorf("state = %v, want stateUnrecoverable", s.state)
	}

	got := s.status.Render()

	if !strings.Contains(got, "/opt/farental/.Farental.old") {
		t.Errorf("status = %q, want it to name the saved old binary path", got)
	}

	if strings.Contains(got, lokyn.L("Press r to retry.")) {
		t.Errorf("status = %q, must not suggest retrying: the target is gone", got)
	}
}

// stateUnrecoverable must behave like stateManualRequired for the exit key:
// escapable when the client was only optionally updating, not otherwise —
// and, unlike stateFailed, there is deliberately no 'r' case at all.
func TestHandleKeyUnrecoverableEsc(t *testing.T) {
	t.Run("optional mode reaches login", func(t *testing.T) {
		s := &Screen{state: stateUnrecoverable, result: updater.Result{Mode: updater.ModeOptional}}

		_, handled := s.handleKey(escKeyMsg())

		if !handled {
			t.Fatal("expected esc to be handled in stateUnrecoverable for ModeOptional")
		}

		if got := orvyn.GetCurrentScreenID(); got != screen.IDLogin {
			t.Errorf("current screen = %q, want %q", got, screen.IDLogin)
		}
	})

	t.Run("mandatory mode has no exit", func(t *testing.T) {
		before := orvyn.GetCurrentScreenID()

		s := &Screen{state: stateUnrecoverable, result: updater.Result{Mode: updater.ModeMandatory}}

		cmd, handled := s.handleKey(escKeyMsg())

		if handled {
			t.Error("esc must not be handled in stateUnrecoverable for ModeMandatory")
		}

		if cmd != nil {
			t.Error("expected no command for an unhandled key")
		}

		if got := orvyn.GetCurrentScreenID(); got != before {
			t.Errorf("current screen changed to %q, want unchanged %q", got, before)
		}
	})

	t.Run("r does not retry", func(t *testing.T) {
		s := &Screen{state: stateUnrecoverable, result: updater.Result{Mode: updater.ModeOptional}}

		cmd, handled := s.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})

		if handled {
			t.Error("'r' must not be handled in stateUnrecoverable: the target binary is gone")
		}

		if cmd != nil {
			t.Error("expected no command for an unhandled key")
		}
	})
}

// OnEnter runs synchronously from App.Init(), before bubbletea's loop starts
// and so before the first tea.WindowSizeMsg arrives — refreshNotes there
// wraps against orvyn's hardcoded startup default rather than the real
// terminal size. Update must handle tea.WindowSizeMsg itself so the notes
// pane is (re)populated once the real size is known, both at startup and on
// every later resize.
func TestUpdateWindowSizeMsgRefreshesNotes(t *testing.T) {
	s := New()
	s.result = updater.Result{
		Notes: []updater.Block{{Type: "p", Spans: []updater.Span{{Text: "hello there"}}}},
	}

	s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	got := s.notes.GetContent()

	if len(got) == 0 {
		t.Fatal("expected a tea.WindowSizeMsg to populate the notes pane")
	}

	if joined := strings.Join(got, " "); !strings.Contains(joined, "hello there") {
		t.Errorf("content = %q, want it to contain the release notes text", joined)
	}
}

// The subtitle must not read "1.1.0  →  " with a blank right side when the
// manifest fetch failed and Latest was never populated.
func TestOnEnterSkipsSubtitleWhenLatestIsEmpty(t *testing.T) {
	s := New()
	updater.Pending = updater.Result{Mode: updater.ModeMandatory, Current: "1.1.0", Err: errors.New("network down")}

	s.OnEnter(nil)

	if got := s.subtitle.Render(); got != "" {
		t.Errorf("subtitle = %q, want empty when Latest is unset", got)
	}
}
