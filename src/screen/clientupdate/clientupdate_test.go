package clientupdate

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"farental/internal/config"
	"farental/internal/keybind"
	"farental/internal/style"
	"farental/internal/updater"
	"farental/screen"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
)

// dummyScreen is a no-op orvyn.Screen, registered under screen.IDLogin (and,
// for user-initiated-check tests, screen.IDUserSettings, screen.IDDashBoard
// and screen.IDClientUpdate) so that handleKey's and OpenCheck's
// orvyn.SwitchScreen calls have somewhere real to go: an unregistered
// ScreenID makes orvyn.SwitchScreen call log.Fatalf, which exits the test
// process outright.
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
	bubblehelp.Init()

	orvyn.Init()

	// Mirrors main.go's own init order: keybind.InitContexts registers
	// ContextClientUpdate (and every other context) with bubblehelp, which
	// TestRefreshHelpKeysEnterVisibility and TestRefreshHelpKeysEscLabel
	// below need to actually be present in bubblehelp.Contexts rather than
	// silently no-op against an unregistered context.
	style.InitHelpStyle()
	keybind.InitContexts()

	orvyn.RegisterScreen(screen.IDLogin, dummyScreen{})
	orvyn.RegisterScreen(screen.IDUserSettings, dummyScreen{})
	orvyn.RegisterScreen(screen.IDClientUpdate, dummyScreen{})
	orvyn.RegisterScreen(screen.IDDashBoard, dummyScreen{})

	os.Exit(m.Run())
}

func fileInfo() updater.FileInfo {
	return updater.FileInfo{URL: "farental-linux-amd64", SHA256: "deadbeef"}
}

// TestDecideEntry covers the single pure decision decideEntry makes, on both
// the startup path (checkErr always nil, since main.go's own fetch has
// already happened by the time enterPending calls this) and a user-initiated
// check's path (checkErr set when checkForUpdates's own request failed).
func TestDecideEntry(t *testing.T) {
	tests := []struct {
		name         string
		result       updater.Result
		checkErr     error
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
			// Latest/ServerCompat are set to a compatible pair so the
			// reasonNoCompatibleRelease check (which runs first) passes
			// through, isolating the no-file case this test is about. See
			// TestDecideEntryNoCompatibleRelease for the mismatched case.
			name: "mandatory mode no file goes manual required",
			result: updater.Result{
				Mode: updater.ModeMandatory, Latest: "1.2.0", ServerCompat: "1.2",
			},
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
			name:       "mandatory mode fetch error goes check failed",
			result:     updater.Result{Mode: updater.ModeMandatory, Err: errors.New("network down")},
			wantState:  stateCheckFailed,
			wantReason: reasonFetchFailed,
		},
		{
			name:       "optional mode happy path goes to prompt",
			result:     updater.Result{Mode: updater.ModeOptional, File: fileInfo()},
			wantState:  statePrompt,
			wantReason: reasonNone,
		},
		{
			// Latest satisfies ServerCompat: the mandatory update this
			// client needs is exactly the one on offer, so the new
			// reasonNoCompatibleRelease check must not fire.
			name: "mandatory mode happy path goes to prompt",
			result: updater.Result{
				Mode: updater.ModeMandatory, Latest: "1.2.0", ServerCompat: "1.2", File: fileInfo(),
			},
			wantState:  statePrompt,
			wantReason: reasonNone,
		},
		{
			// The owner's exact reported defect: client 1.1.0, server
			// demands 1.2, and the newest published release is still 1.1.0.
			// checkAt correctly marks this mandatory (current is
			// incompatible), but the release on offer would not fix
			// anything - updating would just download the same 1.1.0 again.
			name: "mandatory mode, latest release still incompatible (owner's exact case)",
			result: updater.Result{
				Mode: updater.ModeMandatory, Current: "1.1.0", Latest: "1.1.0", ServerCompat: "1.2", File: fileInfo(),
			},
			wantState:  stateManualRequired,
			wantReason: reasonNoCompatibleRelease,
		},
		{
			// A second shape of the same root cause: client 1.0.0, server
			// demands 1.2, latest published is 1.1.0. Updating is real
			// progress (1.0.0 -> 1.1.0) but still does not satisfy the
			// server, so it must not be offered as if it resolves things.
			name: "mandatory mode, latest release is progress but still incompatible",
			result: updater.Result{
				Mode: updater.ModeMandatory, Current: "1.0.0", Latest: "1.1.0", ServerCompat: "1.2", File: fileInfo(),
			},
			wantState:  stateManualRequired,
			wantReason: reasonNoCompatibleRelease,
		},
		{
			// Counterpart to the two cases above: once the latest published
			// release does satisfy ServerCompat, this must not fire - the
			// ordinary prompt is the right presentation.
			name: "mandatory mode, latest release resolves incompatibility goes to prompt",
			result: updater.Result{
				Mode: updater.ModeMandatory, Current: "1.1.0", Latest: "1.2.0", ServerCompat: "1.2", File: fileInfo(),
			},
			wantState:  statePrompt,
			wantReason: reasonNone,
		},
		{
			// ModeOptional must never consult ServerCompat/Latest at all: a
			// compatible client is by definition already satisfied,
			// regardless of what the latest published release happens to
			// be relative to the server's requirement.
			name: "optional mode unaffected by server compat mismatch",
			result: updater.Result{
				Mode: updater.ModeOptional, Current: "1.1.0", Latest: "1.1.0", ServerCompat: "1.2", File: fileInfo(),
			},
			wantState:  statePrompt,
			wantReason: reasonNone,
		},
		{
			// ModeNone likewise skips the new check entirely.
			name: "mode none unaffected by server compat mismatch",
			result: updater.Result{
				Mode: updater.ModeNone, Current: "1.1.0", Latest: "1.1.0", ServerCompat: "1.2",
			},
			wantState:  stateUpToDate,
			wantReason: reasonNone,
		},
		{
			// Only reachable from a user-initiated check: checkForUpdates's
			// own request (the version-compat fetch) never ran at all, so
			// there is no Result to look at yet.
			name:       "check's own request failed",
			checkErr:   errors.New("version endpoint down"),
			wantState:  stateCheckFailed,
			wantReason: reasonFetchFailed,
		},
		{
			// Check's internal manifest fetch failed instead - checkErr is
			// nil, but the Result it did return carries its own Err.
			name:       "check's internal manifest fetch failed",
			result:     updater.Result{Err: errors.New("manifest unreachable")},
			wantState:  stateCheckFailed,
			wantReason: reasonFetchFailed,
		},
		{
			// Only reachable from a user-initiated check: the startup path
			// never calls decideEntry with ModeNone at all (enterPending
			// bails to login first).
			name:       "already up to date",
			result:     updater.Result{Mode: updater.ModeNone, Current: "1.2.0"},
			wantState:  stateUpToDate,
			wantReason: reasonNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotState, gotReason := decideEntry(tt.result, tt.checkErr, tt.preflightErr)

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

	gotState, gotReason := decideEntry(result, nil, nil)

	if gotState != stateCheckFailed {
		t.Errorf("state = %v, want stateCheckFailed", gotState)
	}

	if gotReason != reasonFetchFailed {
		t.Errorf("reason = %v, want %v", gotReason, reasonFetchFailed)
	}
}

// reasonNoCompatibleRelease takes priority over reasonNoFile: if the
// published release cannot resolve the incompatibility, whether a build
// exists for this platform is moot, and reasonNoFile's "no build for your
// platform" wording would misreport a compatibility gap as a
// platform-support one. File is left at its zero value here specifically so
// HasFile() would also fail, proving the compat check is what actually fired.
func TestDecideEntryNoCompatibleReleaseTakesPriorityOverMissingFile(t *testing.T) {
	result := updater.Result{Mode: updater.ModeMandatory, Current: "1.1.0", Latest: "1.1.0", ServerCompat: "1.2"}

	gotState, gotReason := decideEntry(result, nil, nil)

	if gotState != stateManualRequired {
		t.Errorf("state = %v, want stateManualRequired", gotState)
	}

	if gotReason != reasonNoCompatibleRelease {
		t.Errorf("reason = %v, want %v", gotReason, reasonNoCompatibleRelease)
	}
}

// checkErr (checkForUpdates's own request failing) takes priority over
// result.Err (Check's internal manifest fetch failing): without it there is
// no Result to look at, so it must not be shadowed by whatever zero-valued
// Result happens to accompany it.
func TestDecideEntryCheckErrTakesPriorityOverResultErr(t *testing.T) {
	checkErr := errors.New("version endpoint down")
	result := updater.Result{Err: errors.New("manifest unreachable")}

	gotState, gotReason := decideEntry(result, checkErr, nil)

	if gotState != stateCheckFailed {
		t.Errorf("state = %v, want stateCheckFailed", gotState)
	}

	if gotReason != reasonFetchFailed {
		t.Errorf("reason = %v, want reasonFetchFailed", gotReason)
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

// TestHandleKeyCheckFailedEsc covers stateCheckFailed's own esc gating - the
// state re-split from stateManualRequired for a failed check (see decideEntry
// and enterCheckFailed): a compatible client (ModeOptional, startup path)
// must be able to leave with esc, a mandatory one must not, and a
// user-initiated check must always be able to leave regardless of what the
// (failed) check reported - even a Mode it never actually confirmed.
func TestHandleKeyCheckFailedEsc(t *testing.T) {
	t.Run("optional mode reaches login", func(t *testing.T) {
		s := &Screen{state: stateCheckFailed, result: updater.Result{Mode: updater.ModeOptional}}

		_, handled := s.handleKey(escKeyMsg())

		if !handled {
			t.Fatal("expected esc to be handled in stateCheckFailed for ModeOptional")
		}

		if got := orvyn.GetCurrentScreenID(); got != screen.IDLogin {
			t.Errorf("current screen = %q, want %q", got, screen.IDLogin)
		}
	})

	t.Run("mandatory mode has no exit", func(t *testing.T) {
		before := orvyn.GetCurrentScreenID()

		s := &Screen{state: stateCheckFailed, result: updater.Result{Mode: updater.ModeMandatory}}

		cmd, handled := s.handleKey(escKeyMsg())

		if handled {
			t.Error("esc must not be handled in stateCheckFailed for ModeMandatory")
		}

		if cmd != nil {
			t.Error("expected no command for an unhandled key")
		}

		if got := orvyn.GetCurrentScreenID(); got != before {
			t.Errorf("current screen changed to %q, want unchanged %q", got, before)
		}
	})

	t.Run("user-initiated always escapes even when mandatory", func(t *testing.T) {
		orvyn.SwitchScreen(screen.IDUserSettings)

		s := &Screen{
			state:         stateCheckFailed,
			userInitiated: true,
			checkFrom:     screen.IDUserSettings,
			result:        updater.Result{Mode: updater.ModeMandatory},
		}

		_, handled := s.handleKey(escKeyMsg())

		if !handled {
			t.Fatal("expected esc to be handled: a user-initiated check always escapes")
		}

		if got := orvyn.GetCurrentScreenID(); got != screen.IDUserSettings {
			t.Errorf("current screen = %q, want %q (s.checkFrom)", got, screen.IDUserSettings)
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

	if got := s.statusMessage.Render(); !strings.Contains(got, lokyn.L("Press r to retry.")) {
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

	got := s.statusMessage.Render()

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

// By the time handleFinished sees a nil error the swap has already happened,
// but restarting is the user's call: the screen holds at stateInstalled and
// only hands main the go-ahead once enter is pressed, rather than yanking the
// terminal away while they are still reading.
func TestHandleFinishedWaitsForEnterBeforeRestarting(t *testing.T) {
	updater.RestartPending = false
	t.Cleanup(func() { updater.RestartPending = false })

	s := New()
	s.state = stateApplying

	// startUpdate turns the bar on when the download begins; this test enters
	// the state directly, so stand that precondition up by hand - the point
	// below is that handleFinished must not turn it back off.
	s.bar.SetActive(true)

	s.handleFinished(finishedMsg{})

	if s.state != stateInstalled {
		t.Errorf("state = %v, want stateInstalled", s.state)
	}

	// refreshProgress only runs while downloading, and only once a second, so
	// a download finishing inside a tick would otherwise leave the bar at the
	// 0% startUpdate set. Completion has to fill it.
	if got := s.bar.Percent(); got != 1 {
		t.Errorf("progress bar target = %v, want 1 (a completed download reads 100%%)", got)
	}

	if updater.RestartPending {
		t.Error("RestartPending set before the user asked to restart")
	}

	// The completed bar is the only on-screen evidence the download finished,
	// so it stays up while the user decides to restart.
	if !s.bar.IsActive() {
		t.Error("progress bar hidden in stateInstalled; it should stay visible at 100%")
	}

	cmd, handled := s.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	if !handled {
		t.Fatal("enter was not handled in stateInstalled")
	}

	if !updater.RestartPending {
		t.Error("RestartPending not set after enter; main would never exec the new binary")
	}

	if cmd == nil {
		t.Error("expected a quit command so main can exec once bubbletea has torn down")
	}

	if s.state != stateRestarting {
		t.Errorf("state = %v, want stateRestarting", s.state)
	}
}

// When no published release can satisfy the server, the subtitle must name the
// version the server actually requires. The generic subtitle set before
// decideEntry runs reads "current -> latest published", which in this case is
// the very version that cannot help - for a 1.1.0 client refused by a server
// demanding 1.2 with 1.1.0 still the newest release, that renders the useless
// "1.1.0  ->  1.1.0".
func TestEnterManualNoCompatibleReleaseSubtitleShowsRequiredVersion(t *testing.T) {
	s := New()
	s.result = updater.Result{
		Mode:         updater.ModeMandatory,
		Current:      "1.1.0",
		Latest:       "1.1.0",
		ServerCompat: "1.2",
	}

	// Both callers set this generic subtitle before decideEntry picks the
	// state, so enterManual has to overwrite it rather than merely fill a
	// blank one - reproduce that here or the test passes vacuously.
	s.subtitle.SetValue(fmt.Sprintf("%s  →  %s", s.result.Current, s.result.Latest))

	s.enterManual(reasonNoCompatibleRelease, nil)

	got := ansi.Strip(s.subtitle.Render())

	if !strings.Contains(got, "1.2") {
		t.Errorf("subtitle = %q, want it to name the required version 1.2", got)
	}

	if strings.Contains(got, "1.1.0  →  1.1.0") {
		t.Errorf("subtitle = %q, still points at the latest published version, which cannot resolve the incompatibility", got)
	}
}

// A resize (tea.WindowSizeMsg) must not make the notes pane visible in states
// that deliberately hide it. States like stateDownloading and
// stateManualRequired call SetActive(false) to hide the pane, and any
// subsequent resize must preserve that visibility, not force it visible just
// because there's content to wrap.
func TestUpdateWindowSizeMsgPreservesHiddenNotesInDownloadingState(t *testing.T) {
	s := New()
	s.state = stateDownloading
	s.result = updater.Result{
		Notes: []updater.Block{{Type: "p", Spans: []updater.Span{{Text: "release notes"}}}},
	}

	// Simulate the state transition that hides the notes (as startUpdate does).
	s.notes.SetActive(false)

	// Verify the pane is hidden before the resize.
	if s.notes.IsActive() {
		t.Fatal("setup: expected notes to be inactive after SetActive(false)")
	}

	// Send a resize message, which triggers refreshNotes.
	s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Verify the notes are still hidden, not re-activated by the resize.
	if s.notes.IsActive() {
		t.Error("a resize should not make the notes pane visible in stateDownloading")
	}
}

// A resize must preserve the hidden state in stateManualRequired as well.
func TestUpdateWindowSizeMsgPreservesHiddenNotesInManualRequiredState(t *testing.T) {
	s := New()
	s.state = stateManualRequired
	s.result = updater.Result{
		Notes: []updater.Block{{Type: "p", Spans: []updater.Span{{Text: "release notes"}}}},
	}

	// Simulate the state transition that hides the notes (as enterManual does).
	s.notes.SetActive(false)

	// Verify the pane is hidden before the resize.
	if s.notes.IsActive() {
		t.Fatal("setup: expected notes to be inactive after SetActive(false)")
	}

	// Send a resize message.
	s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Verify the notes are still hidden.
	if s.notes.IsActive() {
		t.Error("a resize should not make the notes pane visible in stateManualRequired")
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

// TestOnEnterUserInitiatedConsumesCheckPendingFlag exercises the actual
// hand-off mechanism: OpenCheck cannot pass a parameter through OnEnter (see
// the package doc on checkPending), so OnEnter must read the package-level
// flag and clear it immediately, and must not touch updater.Pending at all.
func TestOnEnterUserInitiatedConsumesCheckPendingFlag(t *testing.T) {
	checkPending = true
	checkFrom = screen.IDUserSettings
	checkRestorePrevious = screen.IDDashBoard
	updater.Pending = updater.Result{Mode: updater.ModeMandatory, Current: "unrelated-pending-value"}

	s := New()
	cmd := s.OnEnter(nil)

	if checkPending {
		t.Error("expected OnEnter to consume checkPending")
	}

	if !s.userInitiated {
		t.Error("expected s.userInitiated to be true")
	}

	if s.checkFrom != screen.IDUserSettings {
		t.Errorf("s.checkFrom = %q, want %q", s.checkFrom, screen.IDUserSettings)
	}

	if s.checkRestorePrevious != screen.IDDashBoard {
		t.Errorf("s.checkRestorePrevious = %q, want %q", s.checkRestorePrevious, screen.IDDashBoard)
	}

	if s.state != stateChecking {
		t.Errorf("state = %v, want stateChecking", s.state)
	}

	// enterUserInitiated must return the fresh-check tea.Cmd (run later by
	// bubbletea) rather than nil, so the UI is not left hanging forever in
	// stateChecking. The command itself talks to the network, so it is
	// never invoked here.
	if cmd == nil {
		t.Error("expected OnEnter to return the fresh-check command, not nil")
	}

	if s.result.Current == "unrelated-pending-value" {
		t.Error("a user-initiated check must not reuse updater.Pending")
	}
}

// A plain OnEnter (checkPending left unset) must still take the startup
// path, unaffected by the user-initiated check path existing at all.
func TestOnEnterWithoutCheckPendingTakesStartupPath(t *testing.T) {
	checkPending = false
	updater.Pending = updater.Result{Mode: updater.ModeOptional, Current: "1.1.0", File: fileInfo()}

	s := New()
	s.OnEnter(nil)

	if s.userInitiated {
		t.Error("expected s.userInitiated to be false")
	}

	if s.state != statePrompt {
		t.Errorf("state = %v, want statePrompt", s.state)
	}
}

// TestEnterPendingBugStopsWhenPendingModeIsNone covers the Critical 1
// defensive guard: main only ever switches to this screen on the startup
// path once updater.Pending.Mode != ModeNone, so reaching enterPending with
// ModeNone at all is already a bug elsewhere (the ping-pong through user
// settings' esc handler being exactly how it could happen). The prompt must
// never render for a version that is already current - there would be no
// way out but ctrl+c - so this must bail straight back to login instead.
func TestEnterPendingBugStopsWhenPendingModeIsNone(t *testing.T) {
	checkPending = false
	updater.Pending = updater.Result{Mode: updater.ModeNone, Current: "1.2.0", Latest: "1.2.0", File: fileInfo()}

	s := New()
	s.OnEnter(nil)

	if got := orvyn.GetCurrentScreenID(); got != screen.IDLogin {
		t.Errorf("current screen = %q, want %q; a ModeNone startup entry must never render the update prompt",
			got, screen.IDLogin)
	}
}

// TestOpenCheckSwitchesToClientUpdateScreen checks the exported entry point
// itself: it must set checkPending, record the caller as checkFrom, capture
// whatever previousScreenID held before its own SwitchScreen call overwrites
// it, and perform the screen switch - the same hand-off shape updater.Pending
// and session.Expired() already use elsewhere in this codebase.
func TestOpenCheckSwitchesToClientUpdateScreen(t *testing.T) {
	defer func() { checkPending = false }()

	// dashboard -> usersettings, as dashboard.go's own handleKey does:
	// previousScreenID becomes screen.IDDashBoard, which OpenCheck must
	// capture into checkRestorePrevious before its own SwitchScreen call
	// below overwrites it with screen.IDUserSettings.
	orvyn.SwitchScreen(screen.IDDashBoard)
	orvyn.SwitchScreen(screen.IDUserSettings)

	OpenCheck(screen.IDUserSettings)

	if got := orvyn.GetCurrentScreenID(); got != screen.IDClientUpdate {
		t.Errorf("current screen = %q, want %q", got, screen.IDClientUpdate)
	}

	// The dummy registered under screen.IDClientUpdate for this test does
	// not consume the flag the way the real Screen.OnEnter does, so it
	// should still read true here.
	if !checkPending {
		t.Error("expected checkPending to be set")
	}

	if checkFrom != screen.IDUserSettings {
		t.Errorf("checkFrom = %q, want %q", checkFrom, screen.IDUserSettings)
	}

	if checkRestorePrevious != screen.IDDashBoard {
		t.Errorf("checkRestorePrevious = %q, want %q (captured before SwitchScreen overwrote previousScreenID)",
			checkRestorePrevious, screen.IDDashBoard)
	}
}

// handleChecked drives the three-way outcome once a user-initiated check's
// fresh fetch comes back: newer available reuses statePrompt (and so the
// same startUpdate/Enter flow), up to date and a failed check each land in
// their own presentation (stateUpToDate, and the ordinary stateManualRequired
// respectively).
func TestHandleCheckedNewerAvailable(t *testing.T) {
	s := New()

	cmd := s.handleChecked(checkedMsg{
		result: updater.Result{Mode: updater.ModeOptional, Current: "1.2.0", Latest: "1.3.0", File: fileInfo()},
	})

	if cmd != nil {
		t.Errorf("expected no command, got %v", cmd)
	}

	if s.state != statePrompt {
		t.Errorf("state = %v, want statePrompt", s.state)
	}
}

func TestHandleCheckedUpToDate(t *testing.T) {
	s := New()

	cmd := s.handleChecked(checkedMsg{
		result: updater.Result{Mode: updater.ModeNone, Current: "1.2.0"},
	})

	if cmd != nil {
		t.Errorf("expected no command, got %v", cmd)
	}

	if s.state != stateUpToDate {
		t.Errorf("state = %v, want stateUpToDate", s.state)
	}

	if got := s.subtitle.Render(); got != "1.2.0" {
		t.Errorf("subtitle = %q, want the current version alone", got)
	}
}

// TestEnterManualTitleAndStatus covers stateManualRequired's title/status
// matrix: reasonNoFile and reasonNotWritable each split on whether the
// client is actually incompatible (result.Mode == updater.ModeMandatory,
// known for certain here since decideEntry only reaches either reason once
// the check itself has already succeeded), independently of the check's
// own escapability.
func TestEnterManualTitleAndStatus(t *testing.T) {
	tests := []struct {
		name       string
		mode       updater.Mode
		reason     entryReason
		wantTitle  string
		wantStatus string
	}{
		{
			name:       "no file, incompatible",
			mode:       updater.ModeMandatory,
			reason:     reasonNoFile,
			wantTitle:  lokyn.L("Version not compatible"),
			wantStatus: lokyn.L("No build is published for your platform, so updating is not possible."),
		},
		{
			name:       "no file, compatible",
			mode:       updater.ModeOptional,
			reason:     reasonNoFile,
			wantTitle:  lokyn.L("Update available"),
			wantStatus: lokyn.L("No build is published for your platform."),
		},
		{
			name:       "not writable, incompatible",
			mode:       updater.ModeMandatory,
			reason:     reasonNotWritable,
			wantTitle:  lokyn.L("Update required"),
			wantStatus: lokyn.L("Farental cannot write to its own directory."),
		},
		{
			name:       "not writable, compatible",
			mode:       updater.ModeOptional,
			reason:     reasonNotWritable,
			wantTitle:  lokyn.L("Update available"),
			wantStatus: lokyn.L("Farental cannot write to its own directory."),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New()
			s.result = updater.Result{Mode: tt.mode}

			s.enterManual(tt.reason, errors.New("read-only filesystem"))

			if s.state != stateManualRequired {
				t.Errorf("state = %v, want stateManualRequired", s.state)
			}

			if got := s.title.Render(); got != tt.wantTitle {
				t.Errorf("title = %q, want %q", got, tt.wantTitle)
			}

			if got := s.statusMessage.Render(); !strings.Contains(got, tt.wantStatus) {
				t.Errorf("status = %q, want it to contain %q", got, tt.wantStatus)
			}

			if got := s.detail.Render(); !strings.Contains(got, config.WebURL) {
				t.Errorf("detail = %q, want the download URL", got)
			}
		})
	}
}

// TestEnterManualNoCompatibleRelease covers reasonNoCompatibleRelease: the
// server's compatibility requirement and the latest published release are
// both named in the status message, severity is error (nothing the user can
// do fixes this - the wording reads as "wait", not "go do something"), and -
// unlike reasonNoFile/reasonNotWritable - no download link is ever shown:
// no available download resolves this, so pointing at the page would be the
// same false advice already removed from enterCheckFailed's
// unreachable-server case. Escapability still comes from canEscape() alone.
func TestEnterManualNoCompatibleRelease(t *testing.T) {
	t.Run("startup path, mandatory, not escapable", func(t *testing.T) {
		s := New()
		s.result = updater.Result{
			Mode: updater.ModeMandatory, Current: "1.1.0", Latest: "1.1.0", ServerCompat: "1.2",
		}

		s.enterManual(reasonNoCompatibleRelease, nil)

		if s.state != stateManualRequired {
			t.Errorf("state = %v, want stateManualRequired", s.state)
		}

		if got := s.title.Render(); got != lokyn.L("Version not compatible") {
			t.Errorf("title = %q, want %q", got, lokyn.L("Version not compatible"))
		}

		got := s.statusMessage.Render()

		wantMsg := fmt.Sprintf(lokyn.L("The server requires version %s, but the latest published version is %s."),
			"1.2", "1.1.0")

		if !strings.Contains(got, wantMsg) {
			t.Errorf("status = %q, want it to contain %q", got, wantMsg)
		}

		if !strings.Contains(got, lokyn.L("Updating is not possible yet.")) {
			t.Errorf("status = %q, want it to contain %q", got, lokyn.L("Updating is not possible yet."))
		}

		detail := s.detail.Render()

		if strings.Contains(detail, config.WebURL) {
			t.Errorf("detail = %q, must not show a download link: no release can resolve this", detail)
		}

		if detail != "" {
			t.Errorf("detail = %q, want empty: not escapable and no download link to show", detail)
		}
	})

	t.Run("user-initiated, escapable, still no download link", func(t *testing.T) {
		s := New()
		s.userInitiated = true
		s.result = updater.Result{
			Mode: updater.ModeMandatory, Current: "1.0.0", Latest: "1.1.0", ServerCompat: "1.2",
		}

		s.enterManual(reasonNoCompatibleRelease, nil)

		detail := s.detail.Render()

		if strings.Contains(detail, config.WebURL) {
			t.Errorf("detail = %q, must not show a download link even when escapable", detail)
		}

		if !strings.Contains(detail, lokyn.L("Press esc to continue without updating.")) {
			t.Errorf("detail = %q, want the esc hint since a user-initiated check is always escapable", detail)
		}
	})
}

// TestHandleCheckedFetchFailed covers stateCheckFailed's "client compatible"
// branch (canEscape is true, here because the check is user-initiated): a
// check that never got a usable Result - checkForUpdates's own request
// failed - must not be presented as "update required", since it never
// learned whether an update exists at all. It gets its own title, a warning
// naming the underlying error (unlike stateManualRequired's fixed wording,
// this one is meant to surface the actual network failure), no download
// link (the host that just failed is not a useful destination), and - since
// a user-initiated check is always escapable - the esc hint.
func TestHandleCheckedFetchFailed(t *testing.T) {
	s := New()
	s.userInitiated = true

	cmd := s.handleChecked(checkedMsg{err: errors.New("network down")})

	if cmd != nil {
		t.Errorf("expected no command, got %v", cmd)
	}

	if s.state != stateCheckFailed {
		t.Errorf("state = %v, want stateCheckFailed", s.state)
	}

	if got := s.title.Render(); got != lokyn.L("Could not check for updates") {
		t.Errorf("title = %q, want %q", got, lokyn.L("Could not check for updates"))
	}

	got := s.statusMessage.Render()

	if !strings.Contains(got, lokyn.L("Could not reach the server.")) {
		t.Errorf("status = %q, want the could-not-reach-the-server message", got)
	}

	if strings.Contains(got, lokyn.L("No build is published for your platform.")) {
		t.Errorf("status = %q, must not misreport a fetch failure as a platform-support gap", got)
	}

	if !strings.Contains(got, "network down") {
		t.Errorf("status = %q, want the underlying error text", got)
	}

	detail := s.detail.Render()

	if strings.Contains(detail, config.WebURL) {
		t.Errorf("detail = %q, must not show a download link: the host that just failed is not a useful destination", detail)
	}

	if !strings.Contains(detail, lokyn.L("Press esc to continue without updating.")) {
		t.Errorf("detail = %q, want the esc hint since a user-initiated check is always escapable", detail)
	}
}

// TestHandleCheckedFetchFailedIncompatible covers stateCheckFailed's other
// branch: canEscape is false because this is the non-user-initiated
// (startup) path and the client is mandatory. Unlike the compatible branch,
// the version-compat fetch itself already succeeded before the manifest
// fetch failed (see updater.checkAt), so Mode is known for certain here -
// and the download URL is this client's only recourse once the host comes
// back, so it must show even though there is no error text and no esc hint.
func TestHandleCheckedFetchFailedIncompatible(t *testing.T) {
	s := New()
	s.userInitiated = false

	cmd := s.handleChecked(checkedMsg{
		result: updater.Result{Mode: updater.ModeMandatory, Err: errors.New("manifest unreachable")},
	})

	if cmd != nil {
		t.Errorf("expected no command, got %v", cmd)
	}

	if s.state != stateCheckFailed {
		t.Errorf("state = %v, want stateCheckFailed", s.state)
	}

	if got := s.title.Render(); got != lokyn.L("Version not compatible") {
		t.Errorf("title = %q, want %q", got, lokyn.L("Version not compatible"))
	}

	got := s.statusMessage.Render()

	wantMsg := lokyn.L("Your version is not compatible with the server, and the server cannot be reached right now.")
	if !strings.Contains(got, wantMsg) {
		t.Errorf("status = %q, want it to contain %q", got, wantMsg)
	}

	detail := s.detail.Render()

	if !strings.Contains(detail, config.WebURL) {
		t.Errorf("detail = %q, want the download URL: it is the only recourse once the host is back", detail)
	}

	if strings.Contains(detail, lokyn.L("Press esc to continue without updating.")) {
		t.Errorf("detail = %q, must not show the esc hint: there is no exit here", detail)
	}
}

// TestHandleCheckedIgnoresStaleGeneration covers Important 2: a
// checkForUpdates command still in flight when the screen is left and
// re-entered - here, on the startup path, before it returns - must not be
// able to rewrite state a newer OnEnter call already set up. Without the
// generation guard, this stale message would flip a perfectly ordinary
// startup-path statePrompt into a user-initiated state whose esc handler
// jumps to login rather than the startup path's own exit.
func TestHandleCheckedIgnoresStaleGeneration(t *testing.T) {
	checkPending = true
	s := New()
	s.OnEnter(nil) // user-initiated entry; generation bumped once.

	staleGen := s.generation

	// The user leaves and the screen is re-entered on the ordinary startup
	// path before the in-flight check (tied to staleGen) comes back.
	checkPending = false
	updater.Pending = updater.Result{Mode: updater.ModeOptional, Current: "1.1.0", File: fileInfo()}
	s.OnEnter(nil) // generation bumped again; staleGen is now out of date.

	if s.userInitiated {
		t.Fatal("setup: expected the second OnEnter to take the startup path")
	}

	if s.state != statePrompt {
		t.Fatalf("setup: state = %v, want statePrompt", s.state)
	}

	cmd := s.handleChecked(checkedMsg{
		gen:    staleGen,
		result: updater.Result{Mode: updater.ModeNone, Current: "9.9.9"},
	})

	if cmd != nil {
		t.Errorf("expected no command for a stale message, got %v", cmd)
	}

	if s.state != statePrompt {
		t.Errorf("a stale checkedMsg hijacked state: state = %v, want unchanged statePrompt", s.state)
	}

	if s.userInitiated {
		t.Error("a stale checkedMsg must not flip the startup path into the user-initiated path")
	}
}

// TestHandleCheckedAppliesCurrentGeneration is
// TestHandleCheckedIgnoresStaleGeneration's counterpart: a message tagged
// with the screen's *current* generation must still be applied normally, so
// the generation guard rejects only genuinely stale messages.
func TestHandleCheckedAppliesCurrentGeneration(t *testing.T) {
	checkPending = true
	s := New()
	s.OnEnter(nil)

	cmd := s.handleChecked(checkedMsg{
		gen:    s.generation,
		result: updater.Result{Mode: updater.ModeNone, Current: "1.2.0"},
	})

	if cmd != nil {
		t.Errorf("expected no command, got %v", cmd)
	}

	if s.state != stateUpToDate {
		t.Errorf("state = %v, want stateUpToDate", s.state)
	}
}

// TestHandleKeyUserInitiatedEscAlwaysExits is the crux of remark B's exit
// rule: esc must return to whichever screen opened the check regardless of
// what the check reported, including a mandatory-update result (possible if
// the server's client_tui compat string changed mid-session) that would
// block esc entirely on the startup path. It must go to s.checkFrom - the
// screen OpenCheck recorded, not whatever orvyn's single previousScreenID
// slot happens to hold (that slot was already overwritten with the caller's
// own ID by OpenCheck's SwitchScreen(IDClientUpdate) call, and relying on it
// instead of checkFrom is exactly the Critical 1 defect: see exitCmd's doc
// comment).
//
// "check failed, mandatory result" covers stateCheckFailed the same way:
// even though a startup entry with this state/mode combination would have
// no exit at all (see TestHandleKeyCheckFailedEsc), a user-initiated one
// must still escape.
func TestHandleKeyUserInitiatedEscAlwaysExits(t *testing.T) {
	tests := []struct {
		name  string
		state state
		mode  updater.Mode
	}{
		{"prompt, mandatory result", statePrompt, updater.ModeMandatory},
		{"prompt, optional result", statePrompt, updater.ModeOptional},
		{"ordinary download failure, mandatory result", stateFailed, updater.ModeMandatory},
		{"check failed, mandatory result", stateCheckFailed, updater.ModeMandatory},
		{"manual required, mandatory result", stateManualRequired, updater.ModeMandatory},
		{"unrecoverable rollback, mandatory result", stateUnrecoverable, updater.ModeMandatory},
		{"still checking", stateChecking, updater.ModeNone},
		{"already up to date", stateUpToDate, updater.ModeNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// A decoy previousScreenID, deliberately different from
			// checkFrom below: if exitCmd read orvyn.GetPreviousScreen()
			// instead of s.checkFrom (the old, broken behavior), the
			// assertion below would catch it landing here instead of on
			// screen.IDUserSettings.
			orvyn.SetPreviousScreen(screen.IDLogin)

			s := &Screen{
				state:         tt.state,
				userInitiated: true,
				checkFrom:     screen.IDUserSettings,
				result:        updater.Result{Mode: tt.mode},
			}

			_, handled := s.handleKey(escKeyMsg())

			if !handled {
				t.Fatalf("expected esc to be handled in the user-initiated path for state %v", tt.state)
			}

			if got := orvyn.GetCurrentScreenID(); got != screen.IDUserSettings {
				t.Errorf("current screen = %q, want %q (s.checkFrom, the recorded opener)", got, screen.IDUserSettings)
			}
		})
	}
}

// TestUserInitiatedExitRestoresPreCheckPreviousScreen reproduces the
// Critical 1 trap end to end: dashboard -> usersettings -> (ctrl+r) check ->
// esc -> usersettings -> esc must land back on dashboard. Before the fix,
// step 3's esc used orvyn.SwitchToPreviousScreen, which orvyn's own
// SwitchScreen(IDClientUpdate) call (in OpenCheck) had already pointed at
// usersettings - so step 3 did go to usersettings, but also left
// previousScreenID pointing at this screen (clientupdate, the screen esc was
// just leaving), since SwitchScreen always records the screen it switched
// *from*. Step 4's esc in user settings is a bare
// orvyn.SwitchToPreviousScreen(), so it re-entered clientupdate instead of
// reaching dashboard - with checkPending now false, landing on the startup
// path with a stale, already-current updater.Pending: exactly the trapped
// state this whole fix exists to prevent.
func TestUserInitiatedExitRestoresPreCheckPreviousScreen(t *testing.T) {
	defer func() { checkPending = false }()

	// Step 1: dashboard -> usersettings, as dashboard.go's own handleKey
	// does; previousScreenID becomes screen.IDDashBoard.
	orvyn.SwitchScreen(screen.IDDashBoard)
	orvyn.SwitchScreen(screen.IDUserSettings)

	// Step 2: ctrl+r from user settings opens the check, exactly as
	// usersettings.go's Update wires it up.
	OpenCheck(screen.IDUserSettings)

	if got := orvyn.GetCurrentScreenID(); got != screen.IDClientUpdate {
		t.Fatalf("setup: current screen = %q, want %q", got, screen.IDClientUpdate)
	}

	// screen.IDClientUpdate is registered to the no-op dummyScreen in
	// TestMain, so it does not consume checkPending/checkFrom the way the
	// real Screen does; build one directly and let it do so, as the other
	// user-initiated tests in this file already do.
	s := New()
	s.OnEnter(nil)

	if s.checkFrom != screen.IDUserSettings {
		t.Fatalf("setup: s.checkFrom = %q, want %q", s.checkFrom, screen.IDUserSettings)
	}

	// Step 3: esc out of the check.
	_, handled := s.handleKey(escKeyMsg())

	if !handled {
		t.Fatal("expected esc to be handled in the user-initiated path")
	}

	if got := orvyn.GetCurrentScreenID(); got != screen.IDUserSettings {
		t.Fatalf("current screen = %q, want %q", got, screen.IDUserSettings)
	}

	if got := orvyn.GetPreviousScreen(); got != screen.IDDashBoard {
		t.Fatalf("previous screen = %q, want %q (restored to what it was before the check)",
			got, screen.IDDashBoard)
	}

	// Step 4: esc in user settings - a bare orvyn.SwitchToPreviousScreen(),
	// exactly as usersettings.go's Update implements it.
	orvyn.SwitchToPreviousScreen()

	if got := orvyn.GetCurrentScreenID(); got != screen.IDDashBoard {
		t.Errorf("current screen = %q, want %q; the user is trapped back in the update screen otherwise",
			got, screen.IDDashBoard)
	}
}

// TestHandleKeyStartupRulesUnchanged fills in the esc-gating cases
// TestHandleKeyManualRequiredEsc and TestHandleKeyUnrecoverableEsc do not
// cover (statePrompt and stateFailed) now that all four states share the
// same canEscape/exitCmd helpers, confirming the startup path's rule -
// escapable only when the update itself is optional - is unchanged by the
// refactor.
func TestHandleKeyStartupRulesUnchanged(t *testing.T) {
	tests := []struct {
		name        string
		state       state
		mode        updater.Mode
		wantHandled bool
	}{
		{"prompt, optional escapes to login", statePrompt, updater.ModeOptional, true},
		{"prompt, mandatory has no exit", statePrompt, updater.ModeMandatory, false},
		{"failed, optional escapes to login", stateFailed, updater.ModeOptional, true},
		{"failed, mandatory has no exit", stateFailed, updater.ModeMandatory, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := orvyn.GetCurrentScreenID()

			s := &Screen{state: tt.state, result: updater.Result{Mode: tt.mode}}

			_, handled := s.handleKey(escKeyMsg())

			if handled != tt.wantHandled {
				t.Errorf("handled = %v, want %v", handled, tt.wantHandled)
			}

			if tt.wantHandled {
				if got := orvyn.GetCurrentScreenID(); got != screen.IDLogin {
					t.Errorf("current screen = %q, want %q", got, screen.IDLogin)
				}

				return
			}

			if got := orvyn.GetCurrentScreenID(); got != before {
				t.Errorf("current screen changed to %q, want unchanged %q", got, before)
			}
		})
	}
}

// TestRefreshHelpKeysEnterVisibility covers the defect from code review:
// ContextClientUpdate advertised "enter: update now" even in states where
// enter is unbound. Enter must be visible only in statePrompt, the only
// state that actually binds it (see handleKey) - stateFailed's retry uses
// 'r', which is deliberately left out of the help keymap entirely (see
// handleFinished), not enter.
func TestRefreshHelpKeysEnterVisibility(t *testing.T) {
	tests := []struct {
		name  string
		state state
		want  bool
	}{
		{"prompt", statePrompt, true},
		{"downloading", stateDownloading, false},
		{"applying", stateApplying, false},
		{"restarting", stateRestarting, false},
		{"failed", stateFailed, false},
		{"manual required", stateManualRequired, false},
		{"check failed", stateCheckFailed, false},
		{"unrecoverable", stateUnrecoverable, false},
		{"checking", stateChecking, false},
		{"up to date", stateUpToDate, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bubblehelp.SwitchContext(keybind.ContextClientUpdate)

			s := &Screen{state: tt.state}
			s.refreshHelpKeys()

			if got := bubblehelp.IsKeybindVisible(keybind.Enter); got != tt.want {
				t.Errorf("enter visible = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRefreshHelpKeysEscLabel covers the esc label half of the same context
// merge: "back" is reserved for the two states unique to a user-initiated
// check, where there is no update on offer to skip; every other state,
// including stateCheckFailed and stateManualRequired, reads "skip".
func TestRefreshHelpKeysEscLabel(t *testing.T) {
	tests := []struct {
		name  string
		state state
		want  string
	}{
		{"prompt reads skip", statePrompt, lokyn.L("skip")},
		{"manual required reads skip", stateManualRequired, lokyn.L("skip")},
		{"check failed reads skip", stateCheckFailed, lokyn.L("skip")},
		{"checking reads back", stateChecking, lokyn.L("back")},
		{"up to date reads back", stateUpToDate, lokyn.L("back")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bubblehelp.SwitchContext(keybind.ContextClientUpdate)

			s := &Screen{state: tt.state}
			s.refreshHelpKeys()

			km := bubblehelp.GetCurrentContextKeymap()

			var got string

			for _, k := range km.Bindings {
				if k.Binding.Help().Key == keybind.Esc.Help().Key {
					got = k.GetHelpDesc()
				}
			}

			if got != tt.want {
				t.Errorf("esc help desc = %q, want %q", got, tt.want)
			}
		})
	}
}
