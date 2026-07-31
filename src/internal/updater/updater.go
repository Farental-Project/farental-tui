package updater

import (
	"farental/internal/config"
	"log"
)

// Mode says what the client must or may do about its version.
type Mode int

const (
	// ModeNone means the client is compatible and up to date, or the check
	// failed harmlessly.
	ModeNone Mode = iota

	// ModeOptional means a newer release exists but this client still works.
	ModeOptional

	// ModeMandatory means the server refuses this client's version.
	ModeMandatory
)

// Result is the outcome of Check, consumed by the clientupdate screen.
type Result struct {
	Mode    Mode
	Current string
	Latest  string

	// ServerCompat is the server's Major.Minor compatibility requirement
	// (db_version.client_tui) that produced Mode. Carrying it lets a
	// consumer ask the question checkAt itself never answers: does Latest,
	// the latest *published* release, actually satisfy this requirement?
	// Without it, a mandatory update whose newest release still doesn't
	// meet the requirement looks identical to one that does - both just say
	// ModeMandatory - and offering the former as if it were a fix loops the
	// user through an update that changes nothing.
	ServerCompat string

	Notes []Block
	File  FileInfo
	Err   error
}

// HasFile reports whether a binary exists for this platform.
func (r Result) HasFile() bool {
	return r.File.URL != "" && r.File.SHA256 != ""
}

// Pending carries the startup check to the update screen, the way
// session.Expired() carries session state to the login screen.
var Pending Result

// RestartPending tells main to exec the new binary once bubbletea has torn
// down the terminal.
var RestartPending bool

// Check compares this client against the server and the published release.
// It never returns an error: a failure is reported through Result.
func Check(currentVersion, serverCompat, lang string) Result {
	return checkAt(config.WebURL, currentVersion, serverCompat, lang)
}

func checkAt(baseURL, currentVersion, serverCompat, lang string) Result {
	// Set unconditionally, here, so every return path below - including the
	// early manifest-failure one - carries it. See Result.ServerCompat.
	result := Result{Current: currentVersion, ServerCompat: serverCompat}

	mandatory := !Compatible(currentVersion, serverCompat)

	m, err := fetchManifest(baseURL, lang)

	if err != nil {
		result.Err = err

		// An incompatible client is refused by the server regardless, so it
		// still stops; a compatible one carries on to the login screen.
		if mandatory {
			result.Mode = ModeMandatory
		}

		log.Println("update check failed:", err)

		return result
	}

	result.Latest = m.Version
	result.Notes = m.NotesBlocks
	result.File = m.Files[PlatformKey()]

	switch {
	case mandatory:
		result.Mode = ModeMandatory
	case Newer(m.Version, currentVersion):
		result.Mode = ModeOptional
	default:
		result.Mode = ModeNone
	}

	return result
}
