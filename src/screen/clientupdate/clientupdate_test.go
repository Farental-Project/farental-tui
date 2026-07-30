package clientupdate

import (
	"errors"
	"os"
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
