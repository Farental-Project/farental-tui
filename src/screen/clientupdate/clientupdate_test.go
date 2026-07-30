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

// dummyScreen is a no-op orvyn.Screen, registered under screen.IDLogin (and,
// for consultation-mode tests, screen.IDUserSettings and
// screen.IDClientUpdate) so that handleKey's and OpenConsultation's
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

	orvyn.Init()
	orvyn.RegisterScreen(screen.IDLogin, dummyScreen{})
	orvyn.RegisterScreen(screen.IDUserSettings, dummyScreen{})
	orvyn.RegisterScreen(screen.IDClientUpdate, dummyScreen{})

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

// TestDecideConsultEntry covers the pure decision consultation mode makes
// once a fresh check has come back (or failed to). Unlike decideEntry, it
// must also recognize "already up to date" as its own state, and a newer
// release still runs through decideEntry's platform/writability checks
// unchanged.
func TestDecideConsultEntry(t *testing.T) {
	tests := []struct {
		name         string
		result       updater.Result
		checkErr     error
		preflightErr error
		wantState    state
		wantReason   entryReason
	}{
		{
			name:       "version-compat fetch failed",
			checkErr:   errors.New("network down"),
			wantState:  stateConsultFailed,
			wantReason: reasonFetchFailed,
		},
		{
			name:       "check's own manifest fetch failed",
			result:     updater.Result{Err: errors.New("manifest unreachable")},
			wantState:  stateConsultFailed,
			wantReason: reasonFetchFailed,
		},
		{
			name:       "already up to date",
			result:     updater.Result{Mode: updater.ModeNone, Current: "1.2.0"},
			wantState:  stateConsultUpToDate,
			wantReason: reasonNone,
		},
		{
			name:       "newer optional release available",
			result:     updater.Result{Mode: updater.ModeOptional, File: fileInfo()},
			wantState:  statePrompt,
			wantReason: reasonNone,
		},
		{
			name: "newer release, but server has become incompatible mid-session",
			// decideConsultEntry must still offer the update (statePrompt),
			// not treat ModeMandatory as a hard block: only handleKey's esc
			// gating differs for consultation, not this decision.
			result:     updater.Result{Mode: updater.ModeMandatory, File: fileInfo()},
			wantState:  statePrompt,
			wantReason: reasonNone,
		},
		{
			name:       "newer release but no file for this platform",
			result:     updater.Result{Mode: updater.ModeOptional},
			wantState:  stateManualRequired,
			wantReason: reasonNoFile,
		},
		{
			name:         "newer release but install directory not writable",
			result:       updater.Result{Mode: updater.ModeOptional, File: fileInfo()},
			preflightErr: errors.New("read-only filesystem"),
			wantState:    stateManualRequired,
			wantReason:   reasonNotWritable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotState, gotReason := decideConsultEntry(tt.result, tt.checkErr, tt.preflightErr)

			if gotState != tt.wantState {
				t.Errorf("state = %v, want %v", gotState, tt.wantState)
			}

			if gotReason != tt.wantReason {
				t.Errorf("reason = %v, want %v", gotReason, tt.wantReason)
			}
		})
	}
}

// A version-compat fetch failure takes priority over Check's own result:
// without it there is no Result to look at, so it must not be shadowed by
// whatever zero-valued Result happens to accompany it.
func TestDecideConsultEntryCheckErrTakesPriorityOverResultErr(t *testing.T) {
	checkErr := errors.New("version endpoint down")
	result := updater.Result{Err: errors.New("manifest unreachable")}

	gotState, gotReason := decideConsultEntry(result, checkErr, nil)

	if gotState != stateConsultFailed {
		t.Errorf("state = %v, want stateConsultFailed", gotState)
	}

	if gotReason != reasonFetchFailed {
		t.Errorf("reason = %v, want reasonFetchFailed", gotReason)
	}
}

// TestOnEnterConsultationConsumesPendingFlag exercises the actual hand-off
// mechanism: OpenConsultation cannot pass a parameter through OnEnter (see
// the package doc on consultPending), so OnEnter must read the package-level
// flag and clear it immediately, and must not touch updater.Pending at all.
func TestOnEnterConsultationConsumesPendingFlag(t *testing.T) {
	consultPending = true
	updater.Pending = updater.Result{Mode: updater.ModeMandatory, Current: "unrelated-pending-value"}

	s := New()
	cmd := s.OnEnter(nil)

	if consultPending {
		t.Error("expected OnEnter to consume consultPending")
	}

	if !s.consultation {
		t.Error("expected s.consultation to be true")
	}

	if s.state != stateConsultChecking {
		t.Errorf("state = %v, want stateConsultChecking", s.state)
	}

	// enterConsultation must return the fresh-check tea.Cmd (run later by
	// bubbletea) rather than nil, so the UI is not left hanging forever in
	// stateConsultChecking. The command itself talks to the network, so it
	// is never invoked here.
	if cmd == nil {
		t.Error("expected OnEnter to return the fresh-check command, not nil")
	}

	if s.result.Current == "unrelated-pending-value" {
		t.Error("consultation mode must not reuse updater.Pending")
	}
}

// A plain OnEnter (consultPending left unset) must still take the startup
// path, unaffected by consultation mode existing at all.
func TestOnEnterWithoutConsultPendingTakesStartupPath(t *testing.T) {
	consultPending = false
	updater.Pending = updater.Result{Mode: updater.ModeOptional, Current: "1.1.0", File: fileInfo()}

	s := New()
	s.OnEnter(nil)

	if s.consultation {
		t.Error("expected s.consultation to be false")
	}

	if s.state != statePrompt {
		t.Errorf("state = %v, want statePrompt", s.state)
	}
}

// TestOpenConsultationSwitchesToClientUpdateScreen checks the exported entry
// point itself: it must set consultPending and perform the screen switch, the
// same hand-off shape updater.Pending and session.Expired() already use
// elsewhere in this codebase.
func TestOpenConsultationSwitchesToClientUpdateScreen(t *testing.T) {
	defer func() { consultPending = false }()

	OpenConsultation()

	if got := orvyn.GetCurrentScreenID(); got != screen.IDClientUpdate {
		t.Errorf("current screen = %q, want %q", got, screen.IDClientUpdate)
	}

	// The dummy registered under screen.IDClientUpdate for this test does
	// not consume the flag the way the real Screen.OnEnter does, so it
	// should still read true here.
	if !consultPending {
		t.Error("expected consultPending to be set")
	}
}

// handleConsultChecked drives the three-way consultation outcome once the
// fresh check comes back: newer available reuses statePrompt (and so the
// same startUpdate/Enter flow), up to date and fetch-failed each get their
// own terminal state.
func TestHandleConsultCheckedNewerAvailable(t *testing.T) {
	s := New()

	cmd := s.handleConsultChecked(consultCheckedMsg{
		result: updater.Result{Mode: updater.ModeOptional, Current: "1.2.0", Latest: "1.3.0", File: fileInfo()},
	})

	if cmd != nil {
		t.Errorf("expected no command, got %v", cmd)
	}

	if s.state != statePrompt {
		t.Errorf("state = %v, want statePrompt", s.state)
	}
}

func TestHandleConsultCheckedUpToDate(t *testing.T) {
	s := New()

	cmd := s.handleConsultChecked(consultCheckedMsg{
		result: updater.Result{Mode: updater.ModeNone, Current: "1.2.0"},
	})

	if cmd != nil {
		t.Errorf("expected no command, got %v", cmd)
	}

	if s.state != stateConsultUpToDate {
		t.Errorf("state = %v, want stateConsultUpToDate", s.state)
	}

	if got := s.subtitle.Render(); got != "1.2.0" {
		t.Errorf("subtitle = %q, want the current version alone", got)
	}
}

func TestHandleConsultCheckedFetchFailed(t *testing.T) {
	s := New()

	cmd := s.handleConsultChecked(consultCheckedMsg{err: errors.New("network down")})

	if cmd != nil {
		t.Errorf("expected no command, got %v", cmd)
	}

	if s.state != stateConsultFailed {
		t.Errorf("state = %v, want stateConsultFailed", s.state)
	}

	if got := s.statusMessage.Render(); !strings.Contains(got, "network down") {
		t.Errorf("status = %q, want it to contain the fetch error", got)
	}
}

// TestHandleKeyConsultationEscAlwaysExits is the crux of remark B's exit
// rule: esc must return to whichever screen opened the consultation
// regardless of what the check reported, including a mandatory-update
// result (possible if the server's client_tui compat string changed
// mid-session) that would block esc entirely on the startup path. It must
// go to the *previous* screen (orvyn.SwitchToPreviousScreen), not always to
// login, which is what distinguishes this from the startup path's exit.
func TestHandleKeyConsultationEscAlwaysExits(t *testing.T) {
	tests := []struct {
		name  string
		state state
		mode  updater.Mode
	}{
		{"prompt, mandatory result", statePrompt, updater.ModeMandatory},
		{"prompt, optional result", statePrompt, updater.ModeOptional},
		{"ordinary download failure, mandatory result", stateFailed, updater.ModeMandatory},
		{"manual required, mandatory result", stateManualRequired, updater.ModeMandatory},
		{"unrecoverable rollback, mandatory result", stateUnrecoverable, updater.ModeMandatory},
		{"still checking", stateConsultChecking, updater.ModeNone},
		{"already up to date", stateConsultUpToDate, updater.ModeNone},
		{"consult fetch failed", stateConsultFailed, updater.ModeMandatory},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// screen.IDUserSettings stands in for "whichever screen opened
			// the consultation" here; using it (rather than screen.IDLogin,
			// which is also the startup path's fixed exit target) proves
			// esc actually goes to the *previous* screen and isn't
			// coincidentally landing on the startup path's hardcoded one.
			orvyn.SetPreviousScreen(screen.IDUserSettings)

			s := &Screen{state: tt.state, consultation: true, result: updater.Result{Mode: tt.mode}}

			_, handled := s.handleKey(escKeyMsg())

			if !handled {
				t.Fatalf("expected esc to be handled in consultation mode for state %v", tt.state)
			}

			if got := orvyn.GetCurrentScreenID(); got != screen.IDUserSettings {
				t.Errorf("current screen = %q, want %q (the previous screen)", got, screen.IDUserSettings)
			}
		})
	}
}

// TestHandleKeyStartupRulesUnchanged fills in the esc-gating cases
// TestHandleKeyManualRequiredEsc and TestHandleKeyUnrecoverableEsc do not
// cover (statePrompt and stateFailed) now that all four states share the
// same canEscape/exitCmd helpers, confirming the startup path's rule -
// escapable only when the update itself is optional - is unchanged by the
// consultation-mode refactor.
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
