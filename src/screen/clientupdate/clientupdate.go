package clientupdate

import (
	"errors"
	"farental/core/data/api"
	"farental/core/request"
	"farental/internal/config"
	"farental/internal/helper"
	"farental/internal/keybind"
	ftheme "farental/internal/theme"
	"farental/internal/ticker"
	"farental/internal/updater"
	"farental/screen"
	"farental/widget/help"
	"farental/widget/simplelogviewer"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/progressbar"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/spf13/viper"

	"sync/atomic"
)

type state int

const (
	statePrompt state = iota
	stateDownloading
	stateApplying
	stateRestarting
	stateFailed
	stateManualRequired

	// stateCheckFailed means the check itself never produced a usable
	// Result: checkForUpdates's own request failed (checkErr; only possible
	// on a user-initiated check - see OpenCheck) or updater.Check's internal
	// manifest fetch failed (result.Err; possible on either path). Kept
	// separate from stateManualRequired - which means the check *succeeded*
	// and found an update it merely cannot apply automatically - because a
	// failed check never learned whether an update exists at all: asserting
	// "update required" and linking to a download page on the very host
	// that just failed would both be wrong. See enterCheckFailed and
	// decideEntry.
	stateCheckFailed

	stateUnrecoverable

	// stateChecking and stateUpToDate exist only for a user-initiated check
	// (see OpenCheck): the startup path never has to represent "still
	// checking" or "already up to date" because main only routes to this
	// screen at all when updater.Pending.Mode != ModeNone.
	stateChecking
	stateUpToDate
)

// bytesPerMB is used only for the human-readable size line.
const bytesPerMB = 1024 * 1024

type finishedMsg struct {
	err error
}

// checkedMsg is the result of the fresh, on-demand check a user-initiated
// entry runs (see OpenCheck). err is checkForUpdates's own failure (distinct
// from a populated result whose Err field is set, which is updater.Check's
// internal manifest-fetch failure); either means there is no usable Result
// to show.
//
// gen ties the message back to the specific enterUserInitiated call that
// launched the check that produced it (see Screen.generation). checkForUpdates
// is an in-flight network command; if the screen is left and re-entered - on
// either path - before it returns, a stale message must not be able to
// rewrite whatever the new entry already set up.
type checkedMsg struct {
	result updater.Result
	err    error
	gen    uint64
}

// checkPending tells the next OnEnter to run a user-initiated check: fetch
// fresh data instead of consuming updater.Pending, and let esc leave
// unconditionally. It is the same hand-off shape as updater.Pending and
// session.Expired() - a package-level flag set by the caller and consumed by
// OnEnter - because orvyn.SwitchScreen only ever hands OnEnter the caller's
// OnExit() return value, and login and usersettings both return nil from
// OnExit unconditionally: neither can tell OnEnter "this transition is a
// user-initiated check" any other way.
var checkPending bool

// checkFrom and checkRestorePrevious ride alongside checkPending, carrying
// the extra hand-off OpenCheck needs: the screen esc must return to, and the
// previousScreenID orvyn had before OpenCheck's own SwitchScreen overwrote
// it. See OpenCheck and Screen.exitCmd.
var (
	checkFrom            orvyn.ScreenID
	checkRestorePrevious orvyn.ScreenID
)

// OpenCheck opens the client-update screen for a user-initiated check: it
// always fetches fresh version info rather than reusing updater.Pending,
// shows current/latest version and release notes, and lets the user start
// the update from here if a newer release exists. Unlike the startup path,
// esc always returns to whichever screen called this, even if the check
// comes back reporting a mandatory update - the user opened this voluntarily
// and must never be trapped by it.
//
// from is the ScreenID of the caller. orvyn keeps a single previousScreenID
// slot, not a stack: the SwitchScreen call below is about to overwrite it
// with from, so relying on orvyn.SwitchToPreviousScreen to leave here would
// send esc back into whichever screen last held that slot, not necessarily
// the caller - and if the caller is itself reached via
// SwitchToPreviousScreen (as user settings' own esc handler is), the two
// ping-pong forever. Capturing from explicitly, and capturing what
// previousScreenID was *before* this call so it can be restored on the way
// out, are both required to break that: see exitCmd.
func OpenCheck(from orvyn.ScreenID) tea.Cmd {
	checkPending = true
	checkFrom = from
	checkRestorePrevious = orvyn.GetPreviousScreen()

	return orvyn.SwitchScreen(screen.IDClientUpdate)
}

// Screen drives the whole update: prompt, download, swap, restart. It also
// drives the read-only, user-initiated check entered via OpenCheck - the
// same flow the startup path runs, differing only in who triggered it and
// what happens when the client is already up to date.
type Screen struct {
	title    *orvyn.SimpleRenderable
	subtitle *orvyn.SimpleRenderable

	// detail carries plain, non-alert text the statusMessage widget
	// shouldn't own: the manual-download URL and the "press esc" hint. Those
	// need to stay plainly readable regardless of whatever warning/error is
	// showing in statusMessage above them.
	detail *orvyn.SimpleRenderable

	statusMessage *statusmessage.Widget

	notes *simplelogviewer.Widget

	bar *progressbar.Widget

	help *help.Widget

	layout *layout.CenterLayout

	state state

	// userInitiated is true for the remainder of this screen's lifetime once
	// OnEnter finds checkPending set. It changes esc's behavior (always
	// exits) and where esc goes (back to the caller, not to login).
	userInitiated bool

	// checkFrom and checkRestorePrevious are OnEnter's copy of the
	// package-level checkFrom/checkRestorePrevious vars, taken at the same
	// time and for the same reason s.userInitiated is a copy of checkPending:
	// exitCmd needs them long after OnEnter returns, and the package vars
	// could have been overwritten by then (a second OpenCheck call, from a
	// different screen, while this one is still on screen). See exitCmd.
	checkFrom            orvyn.ScreenID
	checkRestorePrevious orvyn.ScreenID

	result updater.Result

	progress atomic.Int64

	ticker *ticker.Ticker

	// progressCmd carries refreshProgress's returned tea.Cmd from the
	// ticker's onFire callback (which cannot itself return one) to Update's
	// orvyn.TickMsg case, which must batch it in - dropping it stops the
	// progress bar's animation dead.
	progressCmd tea.Cmd

	// generation counts OnEnter calls. It is bumped unconditionally, on both
	// paths, so any checkedMsg tied to an older generation - a check still
	// in flight when the screen was left and re-entered, on either path - is
	// recognized as stale and dropped rather than clobbering whatever the
	// new entry just set up. See handleChecked.
	generation uint64
}

func New() *Screen {
	s := new(Screen)

	t := orvyn.GetTheme()

	s.title = orvyn.NewSimpleRenderable("")
	s.title.Style = t.Style(theme.TitleStyleID)

	s.subtitle = orvyn.NewSimpleRenderable("")
	s.subtitle.Style = t.Style(theme.DimTextStyleID)

	s.detail = orvyn.NewSimpleRenderable("")
	s.detail.Style = t.Style(theme.NormalTextStyleID)

	s.statusMessage = statusmessage.New()

	s.notes = simplelogviewer.New("Release notes")
	s.notes.Style = simplelogviewer.Style{
		FocusedWidget: t.Style(theme.FocusedWidgetStyleID),
		BlurredWidget: t.Style(theme.BlurredWidgetStyleID),
		FocusedTitle:  t.Style(ftheme.TitleUnderlinedTextStyleID),
		BlurredTitle:  t.Style(ftheme.DimUnderlinedTextStyleID),
	}
	s.notes.SetMinSize(orvyn.NewSize(30, 30))
	s.notes.SetPreferredSize(orvyn.NewSize(30, 120))
	s.notes.SetAutoScroll(false)

	// simplelogviewer resolves its border and title styles in OnFocus/OnBlur;
	// without one of them both stay zero-valued and the widget renders bare.
	// This is the only interactive widget on the screen, so it reads focused.
	s.notes.OnFocus()

	s.bar = progressbar.New("")
	s.bar.SetTitleProgressVisibility(false)
	s.bar.SetPercentageVisibility(true)
	s.bar.SetActive(false)

	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(
			35,
			t.Size(ftheme.LayoutWidthSizeID),
			10,
			s.title,
			s.subtitle,
			orvyn.VGap,
			s.notes,
			orvyn.VGap,
			s.bar,
			s.statusMessage,
			s.detail,
			orvyn.VGap,
			s.help,
		),
	)

	s.ticker = ticker.New(1, s.onProgressTick)

	return s
}

func (s *Screen) OnEnter(_ any) tea.Cmd {
	bubblehelp.SwitchContext(keybind.ContextClientUpdate)

	// Bumped unconditionally, before branching, so a checkedMsg from any
	// earlier entry - user-initiated or startup - is stale from this point
	// on. See the doc comment on the generation field.
	s.generation++

	s.userInitiated = checkPending
	checkPending = false

	s.refreshContext()

	s.notes.SetTitle(lokyn.L("release notes"))

	if s.userInitiated {
		s.checkFrom = checkFrom
		s.checkRestorePrevious = checkRestorePrevious

		return s.enterUserInitiated()
	}

	return s.enterPending()
}

// enterPending is the startup path: OnEnter's original behavior, consuming
// updater.Pending exactly as main.go left it.
func (s *Screen) enterPending() tea.Cmd {
	s.result = updater.Pending

	// main only ever switches to this screen on the startup path after
	// finding updater.Pending.Mode != ModeNone; reaching here with ModeNone
	// means something upstream is broken (e.g. a stale re-entry ping-ponging
	// through user settings - see OpenCheck). Rendering the prompt anyway
	// would offer to reinstall the exact version already running, with
	// canEscape false (Mode is never ModeOptional/ModeMandatory here) and so
	// no way out but ctrl+c. Bail to login instead of ever showing that.
	if s.result.Mode == updater.ModeNone {
		return orvyn.SwitchScreen(screen.IDLogin)
	}

	s.setNewUpdateTitle()

	// Latest is empty when the manifest fetch failed; showing "1.1.0  →  "
	// with a blank right side is worse than showing nothing.
	if s.result.Latest != "" {
		s.subtitle.SetValue(fmt.Sprintf("%s  →  %s", s.result.Current, s.result.Latest))
	} else {
		s.subtitle.SetValue("")
	}

	s.refreshNotes()

	s.bar.SetActive(false)
	s.statusMessage.Reset()
	s.detail.SetValue("")

	// PreflightWritable is a cheap local check (create+remove one temp file),
	// so it always runs; decideEntry only consults it once the fetch and
	// platform checks already passed.
	preflightErr := updater.PreflightWritable()

	// checkErr is always nil here: the startup path's fetch already
	// happened, in main.go, before this screen ever opened, so there is
	// nothing of its own for enterPending to fail at.
	entryState, reason := decideEntry(s.result, nil, preflightErr)
	s.state = entryState

	switch entryState {
	case stateCheckFailed:
		s.enterCheckFailed(s.result.Err)
	case stateManualRequired:
		s.enterManual(reason, preflightErr)
	}

	s.refreshHelpKeys()

	return nil
}

// entryReason identifies why OnEnter chose stateCheckFailed or
// stateManualRequired, so OnEnter can pick the right localized message
// without decideEntry depending on lokyn. That keeps decideEntry a pure
// function of its inputs, trivial to unit test.
type entryReason int

const (
	reasonNone entryReason = iota
	reasonFetchFailed

	// reasonNoCompatibleRelease means the check succeeded, the update is
	// mandatory, and the latest *published* release still would not satisfy
	// the server's compatibility requirement - so no build existing for this
	// platform, or the disk being writable, is moot: nothing downloadable
	// right now can fix this client. See decideEntry and enterManual.
	reasonNoCompatibleRelease

	reasonNoFile
	reasonNotWritable
)

// decideEntry is the pure decision made once a check result is in hand -
// updater.Pending on the startup path, or a fresh checkedMsg for a
// user-initiated one (see OpenCheck) - given whatever the preflight check
// already found (so this itself needs no filesystem access). It picks the
// state to enter and, when that state is stateCheckFailed or
// stateManualRequired, why.
//
// checkErr is checkForUpdates's own failure on the user-initiated path -
// distinct from result.Err, which is updater.Check's internal manifest-fetch
// failure - and is always nil on the startup path, where main.go has
// already run the fetch by the time this screen opens. Checked first since
// without it there is no Result to look at. Either failure lands in
// stateCheckFailed, paired with reasonFetchFailed: neither one learned
// whether an update exists at all, which is a different situation from
// stateManualRequired's reasonNoFile/reasonNotWritable, both of which only
// happen once the check itself has already succeeded. See enterCheckFailed.
//
// "already up to date" (result.Mode == updater.ModeNone) only matters for a
// user-initiated check: the startup path never calls this with ModeNone at
// all (enterPending bails to login first - see its own ModeNone guard), so
// that branch is simply never exercised there.
//
// For a mandatory update, whether the latest published release actually
// satisfies the server's requirement is checked next, and deliberately
// before HasFile(): if the release that exists cannot resolve the
// incompatibility anyway, whether a build happens to exist for this
// platform is moot, and reasonNoFile's "no build for your platform" wording
// would misreport a compatibility gap as a platform-support one. ModeOptional
// and ModeNone skip this check entirely - a compatible client is by
// definition already satisfied, whatever Latest happens to be.
//
// Ordered most specific cause first: a failed fetch also leaves File empty
// and Mode at its zero value, so checking either of those before checkErr or
// result.Err would misreport a network problem as a platform-support gap or,
// worse, as "up to date".
func decideEntry(result updater.Result, checkErr, preflightErr error) (state, entryReason) {
	if checkErr != nil {
		return stateCheckFailed, reasonFetchFailed
	}

	if result.Err != nil {
		return stateCheckFailed, reasonFetchFailed
	}

	if result.Mode == updater.ModeNone {
		return stateUpToDate, reasonNone
	}

	if result.Mode == updater.ModeMandatory && !updater.Compatible(result.Latest, result.ServerCompat) {
		return stateManualRequired, reasonNoCompatibleRelease
	}

	if !result.HasFile() {
		return stateManualRequired, reasonNoFile
	}

	if preflightErr != nil {
		return stateManualRequired, reasonNotWritable
	}

	return statePrompt, reasonNone
}

// enterUserInitiated starts the user-initiated flow (see OpenCheck): it
// never touches updater.Pending, always re-fetches, and shows a checking
// state while the fetch (a tea.Cmd, not run inline here) is in flight so the
// UI stays responsive.
func (s *Screen) enterUserInitiated() tea.Cmd {
	s.result = updater.Result{}
	s.progress.Store(0)

	s.state = stateChecking
	s.refreshHelpKeys()

	s.title.SetValue(lokyn.L("Checking for updates..."))
	s.subtitle.SetValue("")
	s.detail.SetValue("")
	s.notes.SetActive(false)
	s.bar.SetActive(false)
	s.statusMessage.Reset()

	// gen is captured now, not read from s inside the closure: by the time
	// this command runs (and, worse, by the time its result comes back),
	// s.generation may have moved on to a later entry.
	gen := s.generation

	return func() tea.Msg {
		msg := checkForUpdates()

		if m, ok := msg.(checkedMsg); ok {
			m.gen = gen
			return m
		}

		return msg
	}
}

// checkForUpdates is a user-initiated check's fresh fetch, run as a tea.Cmd
// so OnEnter returns immediately. It mirrors main.go's startup sequence -
// fetch the server's ClientTui compat string via request.VersionGet, then
// updater.Check - but reports the version-compat fetch's own failure back
// through checkedMsg instead of printing and returning, since there is no
// os.Exit available (or wanted) from inside a running screen.
func checkForUpdates() tea.Msg {
	version, err := helper.Fetch[api.DbVersion](request.VersionGet())

	if err != nil {
		return checkedMsg{err: err}
	}

	result := updater.Check(config.VERSION, version.ClientTui, viper.GetString("language"))

	return checkedMsg{result: result}
}

// handleChecked drives the outcome once a user-initiated check's fresh fetch
// comes back: it mirrors enterPending's own structure (set subtitle/notes
// from the result, then decideEntry, then act on the state it picked) since
// this is the same decision the startup path makes, just reached over a
// tea.Cmd instead of synchronously from updater.Pending.
func (s *Screen) handleChecked(msg checkedMsg) tea.Cmd {
	// checkForUpdates is an in-flight network command; if the screen was
	// left and re-entered - on either path - before it returned, this
	// message belongs to a check OnEnter has already superseded. Applying it
	// now would rewrite s.result/s.state/title/subtitle out from under
	// whatever the new entry set up, including moving a startup-path
	// re-entry into a user-initiated state whose esc handler jumps to login
	// instead of the startup path's own exit.
	if msg.gen != s.generation {
		return nil
	}

	s.result = msg.result
	s.detail.SetValue("")

	// Latest is empty when checkForUpdates's own request failed or Check's
	// internal manifest fetch did; showing "1.1.0  →  " with a blank right
	// side is worse than showing nothing.
	if s.result.Latest != "" {
		s.subtitle.SetValue(fmt.Sprintf("%s  →  %s", s.result.Current, s.result.Latest))
	} else {
		s.subtitle.SetValue("")
	}

	s.refreshNotes()

	preflightErr := updater.PreflightWritable()

	entryState, reason := decideEntry(msg.result, msg.err, preflightErr)
	s.state = entryState

	switch entryState {
	case stateUpToDate:
		s.title.SetValue(lokyn.L("Farental is up to date"))
		s.subtitle.SetValue(msg.result.Current)
		s.statusMessage.SetMessage(lokyn.L("You are running the latest version."), statusmessage.SuccessMessage)

	case statePrompt:
		s.setNewUpdateTitle()
		s.statusMessage.Reset()

	case stateCheckFailed:
		// checkErr (checkForUpdates's own request failing) and result.Err
		// (Check's internal manifest fetch failing) are mutually exclusive -
		// decideEntry only reaches this state via one or the other - so
		// whichever is non-nil is the failure to show.
		err := msg.err
		if err == nil {
			err = msg.result.Err
		}

		s.enterCheckFailed(err)

	case stateManualRequired:
		s.enterManual(reason, preflightErr)
	}

	s.refreshHelpKeys()

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if m, ok := orvyn.GetKeyMsg(msg); ok {
		if cmd, handled := s.handleKey(m); handled {
			return cmd
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Notes are rendered based on the current width of the layout.
		// Needs to be refresh when window size shifts.
		wasActive := s.notes.IsActive()

		s.refreshNotes()

		if s.notes.IsActive() {
			s.notes.SetActive(wasActive)
		}

		return s.notes.Update(msg)

	case progress.FrameMsg:
		return s.bar.Update(msg)

	case orvyn.TickMsg:
		handled, cmd := s.ticker.Handle(msg)

		if !handled {
			return nil
		}

		return tea.Batch(cmd, s.progressCmd)

	case finishedMsg:
		return s.handleFinished(msg)

	case checkedMsg:
		return s.handleChecked(msg)
	}

	return s.notes.Update(msg)
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

// canEscape reports whether esc is allowed to leave the screen from the
// current state: always for a user-initiated check (the user opened this
// voluntarily and must never be trapped by it, even by a mandatory-update
// result), and - for the startup path, unchanged from before - only when the
// update itself is optional. A mandatory update at startup has no exit
// besides quitting.
func (s *Screen) canEscape() bool {
	return s.userInitiated || s.result.Mode == updater.ModeOptional
}

// exitCmd is where esc goes once canEscape allows it: back to whichever
// screen opened a user-initiated check, or to login for the startup path,
// exactly as before.
//
// The user-initiated branch switches to s.checkFrom directly rather than
// orvyn.SwitchToPreviousScreen: orvyn keeps a single previousScreenID slot,
// which OpenCheck's own SwitchScreen(IDClientUpdate) already overwrote with
// the caller's ID, so by now it holds no information exitCmd doesn't already
// have in s.checkFrom. Switching there also overwrites previousScreenID a
// second time - to this screen's own ID - which would otherwise make the
// caller's *next* esc bounce right back here; restoring it to whatever it
// was before the check (captured in s.checkRestorePrevious, see OpenCheck)
// undoes that.
func (s *Screen) exitCmd() tea.Cmd {
	if s.userInitiated {
		cmd := orvyn.SwitchScreen(s.checkFrom)
		orvyn.SetPreviousScreen(s.checkRestorePrevious)

		return cmd
	}

	return orvyn.SwitchScreen(screen.IDLogin)
}

// refreshHelpKeys keeps the help bar's advertised keys honest for the
// current state, within the single ContextClientUpdate context: enter only
// does something in statePrompt (every other state either has no enter
// action at all, or - stateFailed's retry - a key that isn't advertised
// here, see handleFinished), so it is hidden everywhere else. Esc reads
// "back" in the two states unique to a user-initiated check (still checking,
// already up to date), where there is no update on offer to skip, and
// "skip" everywhere else - including stateCheckFailed and stateManualRequired,
// neither of which single themselves out from the ordinary manual-required
// wording.
func (s *Screen) refreshHelpKeys() {
	bubblehelp.SetKeybindVisible(keybind.Enter, s.state == statePrompt)

	switch s.state {
	case stateChecking, stateUpToDate:
		bubblehelp.UpdateKeybindHelpDesc(keybind.Esc, lokyn.L("back"))
	default:
		bubblehelp.UpdateKeybindHelpDesc(keybind.Esc, lokyn.L("skip"))
	}
}

func (s *Screen) handleKey(m tea.KeyMsg) (tea.Cmd, bool) {
	switch s.state {
	case statePrompt:
		switch {
		case key.Matches(m, keybind.Enter):
			return s.startUpdate(), true
		case key.Matches(m, keybind.Esc):
			if s.canEscape() {
				return s.exitCmd(), true
			}
		}

	case stateFailed:
		switch {
		case key.Matches(m, keybind.RKey):
			return s.startUpdate(), true
		case key.Matches(m, keybind.Esc):
			if s.canEscape() {
				return s.exitCmd(), true
			}
		}

	case stateManualRequired, stateCheckFailed:
		switch {
		case key.Matches(m, keybind.Esc):
			if s.canEscape() {
				return s.exitCmd(), true
			}
		}

	case stateUnrecoverable:
		switch {
		case key.Matches(m, keybind.Esc):
			// No retry key here on purpose: the target binary is gone
			// and a retry cannot succeed.
			if s.canEscape() {
				return s.exitCmd(), true
			}
		}

	case stateChecking, stateUpToDate:
		// These two states are only ever entered from enterUserInitiated, so
		// s.userInitiated is always true here: esc always leaves, no
		// canEscape gating needed.
		switch {
		case key.Matches(m, keybind.Esc):
			return s.exitCmd(), true
		}
	}

	return nil, false
}

func (s *Screen) startUpdate() tea.Cmd {
	s.state = stateDownloading
	s.progress.Store(0)

	// A retry from stateFailed left the title reading "Update failed"; a
	// fresh attempt needs it back to the neutral prompt title.
	s.setNewUpdateTitle()

	s.bar.SetActive(true)
	s.notes.SetActive(false)
	s.statusMessage.SetMessage(lokyn.L("Downloading..."), statusmessage.InformationMessage)
	s.detail.SetValue("")

	// statePrompt is the only state that advertises enter; leaving it must
	// hide that key again.
	s.refreshHelpKeys()

	file := s.result.File

	download := func() tea.Msg {
		return finishedMsg{err: updater.Apply(config.WebURL, file, &s.progress)}
	}

	// Restart (not Start) bumps the ticker's tag, invalidating any tick
	// still in flight from a previous attempt - the same thing the old
	// hand-rolled s.tickTag++ did on every call, including a retry.
	//
	// refreshProgress is also called directly, once, right here: Restart
	// only arms the *next* tick after a full interval, it never fires
	// onFire itself, so without this the bar would sit at 0% with no MB
	// counter for the first second of every download.
	return tea.Batch(download, s.refreshProgress(), s.ticker.Restart())
}

// onProgressTick is the ticker's onFire callback. A Ticker's onFire is a
// plain func() with no return value, so refreshProgress's tea.Cmd (which
// must keep reaching bubbletea or the progress bar's animation stalls) is
// stashed here for Update's orvyn.TickMsg case to pick up and batch in.
func (s *Screen) onProgressTick() {
	s.progressCmd = s.refreshProgress()
}

func (s *Screen) refreshProgress() tea.Cmd {
	if s.state != stateDownloading {
		return nil
	}

	done := s.progress.Load()
	total := s.result.File.SizeBytes

	if total <= 0 {
		return nil
	}

	percent := float64(done) / float64(total)

	s.statusMessage.SetMessage(fmt.Sprintf("%.1f / %.1f MB",
		float64(done)/bytesPerMB, float64(total)/bytesPerMB), statusmessage.InformationMessage)

	// The whole body is read before the swap, so a full bar means the
	// download finished and verification is under way.
	if done >= total {
		s.state = stateApplying
		s.statusMessage.SetMessage(lokyn.L("Verifying and installing..."), statusmessage.InformationMessage)
	}

	return s.bar.SetPercent(percent)
}

func (s *Screen) handleFinished(msg finishedMsg) tea.Cmd {
	if msg.err != nil {
		s.bar.SetActive(false)
		s.title.SetValue(lokyn.L("Update failed"))

		// RollbackFailedError means the swap failed *and* selfupdate's own
		// rollback failed: there is no file at the target path at all.
		// Retrying can only fail again the same way, so the hint must not
		// be "press r" — it has to send the user to the saved old binary.
		var rerr *updater.RollbackFailedError

		if errors.As(msg.err, &rerr) {
			s.state = stateUnrecoverable
			s.statusMessage.SetMessage(fmt.Sprintf("%s\n%v\n\n%s",
				lokyn.L("The update failed."), msg.err,
				fmt.Sprintf(lokyn.L("The previous version was saved at %s. Reinstall it manually."), rerr.OldPath)),
				statusmessage.ErrorMessage)

			return nil
		}

		s.state = stateFailed
		// The retry key is not in the help keymap, which is fixed per context,
		// so the hint goes in the message the user is already reading.
		s.statusMessage.SetMessage(fmt.Sprintf("%s\n%v\n\n%s",
			lokyn.L("The update failed."), msg.err, lokyn.L("Press r to retry.")),
			statusmessage.ErrorMessage)

		return nil
	}

	s.state = stateRestarting
	s.statusMessage.SetMessage(lokyn.L("Restarting..."), statusmessage.InformationMessage)

	// The exec happens in main, after bubbletea has left the alt screen and
	// restored the cursor; doing it here would hand the new process a
	// terminal still in raw mode.
	updater.RestartPending = true

	return tea.Quit
}

// enterManual drives stateManualRequired: the check itself succeeded and an
// update exists, but it cannot be applied automatically - because the latest
// published release still would not satisfy the server's requirement
// (reasonNoCompatibleRelease), because no build is published for this
// platform (reasonNoFile), or because Farental cannot write to its own
// directory (reasonNotWritable). All three share the canEscape-gated esc
// hint; reasonNoFile and reasonNotWritable also share the download URL and
// split their title/status-message severity on whether the client is
// actually incompatible. That is s.result.Mode == updater.ModeMandatory,
// known for certain here - unlike in enterCheckFailed - since decideEntry
// only reaches any of these reasons once the check has already succeeded and
// populated Mode for real.
//
// reasonNoCompatibleRelease is always the mandatory/incompatible case -
// decideEntry only returns it for ModeMandatory - and never shows the
// download URL: no available download resolves it, so pointing at the page
// would be the same false advice already removed from enterCheckFailed's
// unreachable-server case. Nothing the user can do here fixes it; the
// wording reads as "wait", not as "go do something".
//
// preflightErr is only used for reasonNotWritable; callers pass it
// unconditionally since both call sites already have it in scope.
func (s *Screen) enterManual(reason entryReason, preflightErr error) {
	s.state = stateManualRequired
	s.bar.SetActive(false)
	s.notes.SetActive(false)

	incompatible := s.result.Mode == updater.ModeMandatory

	showDownloadLink := true

	switch reason {
	case reasonNoCompatibleRelease:
		showDownloadLink = false

		// The subtitle set before decideEntry ran reads "1.1.0  →  1.1.0":
		// the latest published version, which is precisely the one that
		// cannot help here. Point it at what the server actually requires.
		s.subtitle.SetValue(fmt.Sprintf("%s  →  %s", s.result.Current, s.result.ServerCompat))

		s.title.SetValue(lokyn.L("Version not compatible"))
		s.statusMessage.SetMessage(fmt.Sprintf("%s\n%s",
			fmt.Sprintf(lokyn.L("The server requires version %s, but the latest published version is %s."),
				s.result.ServerCompat, s.result.Latest),
			lokyn.L("Updating is not possible yet.")),
			statusmessage.ErrorMessage)

	case reasonNoFile:
		if incompatible {
			s.title.SetValue(lokyn.L("Version not compatible"))
			s.statusMessage.SetMessage(lokyn.L("No build is published for your platform, so updating is not possible."),
				statusmessage.ErrorMessage)
		} else {
			s.title.SetValue(lokyn.L("Update available"))
			s.statusMessage.SetMessage(lokyn.L("No build is published for your platform."), statusmessage.WarningMessage)
		}

	case reasonNotWritable:
		if incompatible {
			s.title.SetValue(lokyn.L("Update required"))
		} else {
			s.title.SetValue(lokyn.L("Update available"))
		}

		s.statusMessage.SetMessage(fmt.Sprintf("%s\n%v",
			lokyn.L("Farental cannot write to its own directory."), preflightErr), statusmessage.WarningMessage)
	}

	// Kept out of statusMessage, which renders one styled (warning- or
	// error-colored) message: the URL needs to stay plainly readable rather
	// than tinted the same as whatever is showing above it, and the hint is
	// guidance, not part of that message itself.
	var detail string

	if showDownloadLink {
		detail = fmt.Sprintf("%s\n%s/clienttui", lokyn.L("Download the new version here:"), config.WebURL)
	}

	if s.canEscape() {
		if detail != "" {
			detail = fmt.Sprintf("%s\n\n%s", detail, lokyn.L("Press esc to continue without updating."))
		} else {
			detail = lokyn.L("Press esc to continue without updating.")
		}
	}

	s.detail.SetValue(detail)
}

// enterCheckFailed drives stateCheckFailed: the check itself never produced
// a usable Result (see decideEntry's reasonFetchFailed branches), so there
// is no way to know whether an update even exists. Unlike enterManual, this
// is not "an update exists but can't be applied": it is "we don't know", and
// the two presentations it can take differ in title, severity and whether
// the download URL is worth showing at all.
//
// The split is on canEscape rather than raw Mode: a user-initiated check
// always escapes regardless of what it reported (see canEscape's own doc
// comment), and in that situation the user never learned this client is
// actually incompatible either - checkForUpdates's own request failing
// leaves Mode at its zero value - so there is nothing more specific than
// "could not check" to tell them. Only the startup path's own mandatory,
// non-escapable case has enough information to assert incompatibility.
func (s *Screen) enterCheckFailed(err error) {
	s.state = stateCheckFailed
	s.bar.SetActive(false)
	s.notes.SetActive(false)

	if s.canEscape() {
		s.title.SetValue(lokyn.L("Could not check for updates"))
		s.statusMessage.SetMessage(fmt.Sprintf("%s\n%v", lokyn.L("Could not reach the server."), err),
			statusmessage.WarningMessage)

		// No download link here: the host that just failed to answer is not
		// a useful place to send the user.
		s.detail.SetValue(lokyn.L("Press esc to continue without updating."))

		return
	}

	// No esc hint: canEscape is false, so there is no esc to hint at. The
	// download URL is this client's only recourse once the host is back.
	s.title.SetValue(lokyn.L("Version not compatible"))
	s.statusMessage.SetMessage(
		lokyn.L("Your version is not compatible with the server, and the server cannot be reached right now."),
		statusmessage.ErrorMessage)
	s.detail.SetValue(fmt.Sprintf("%s\n%s/clienttui", lokyn.L("Download the new version here:"), config.WebURL))
}

func (s *Screen) refreshNotes() {
	lines := updater.RenderNotes(s.result.Notes, s.layout.GetSize().Width)

	if len(lines) == 0 {
		s.notes.SetActive(false)
		return
	}

	s.notes.SetActive(true)
	s.notes.SetContent(lines)
}

func (s *Screen) refreshContext() {
	bubblehelp.SetKeybindVisible(keybind.Esc, s.canEscape())
}

func (s *Screen) setNewUpdateTitle() {
	if s.canEscape() {
		s.title.SetValue(lokyn.L("A new version is available"))
	} else {
		s.title.SetValue(lokyn.L("A new mandatory version is available"))
	}
}
