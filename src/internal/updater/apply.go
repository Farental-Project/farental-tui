package updater

import (
	"context"
	"encoding/hex"
	"farental/core/request"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/minio/selfupdate"
)

// downloadTimeout bounds the whole transfer, not just the connection.
const downloadTimeout = 10 * time.Minute

// maxBinarySize bounds how large a claimed download may be before Apply
// rejects it outright, before any request is made. Current binaries run
// about 22 MB; 256 MB leaves headroom without letting a manifest with a
// bogus (or compromised) size make the client allocate until the OS kills
// it, since selfupdate reads the whole body into memory.
const maxBinarySize = 256 * 1024 * 1024

// targetPathOverride replaces the running executable's path in tests. Empty in
// production, where os.Executable is used.
var targetPathOverride string

// swappedPath is the executable path Apply successfully swapped into place,
// set only once Apply has returned no error. Restart must reuse this rather
// than resolving the path a second time: by the time Restart runs,
// selfupdate has already renamed the pre-update binary aside, so asking the
// OS again either fails (the old inode has no name any more) or, if
// selfupdate's own removal of the saved copy happened to fail, silently
// resolves to the binary the update just replaced.
var (
	swappedPathMu sync.Mutex
	swappedPath   string
)

// rememberSwappedPath records the path Apply just swapped into place.
func rememberSwappedPath(path string) {
	swappedPathMu.Lock()
	swappedPath = path
	swappedPathMu.Unlock()
}

// SwappedExecutablePath returns the path Restart should exec: the path Apply
// swapped into place if an update just succeeded in this process, otherwise
// whatever ExecutablePath resolves right now (the normal case, when no
// update happened this run).
func SwappedExecutablePath() (string, error) {
	swappedPathMu.Lock()
	p := swappedPath
	swappedPathMu.Unlock()

	if p != "" {
		return p, nil
	}

	return ExecutablePath()
}

// isOldBackupName reports whether base has the shape selfupdate (and
// oldBinaryPath, below) use for the saved pre-update binary, e.g.
// ".Farental.old". ExecutablePath must never hand back such a path: if it
// did, Restart could relaunch the binary an update just replaced instead of
// the one it installed.
func isOldBackupName(base string) bool {
	return strings.HasPrefix(base, ".") && strings.HasSuffix(base, ".old")
}

// ExecutablePath returns the path of the binary to replace, with symlinks
// resolved so the real file is swapped rather than a link to it.
func ExecutablePath() (string, error) {
	if targetPathOverride != "" {
		return targetPathOverride, nil
	}

	exe, err := os.Executable()

	if err != nil {
		return "", err
	}

	resolved, err := filepath.EvalSymlinks(exe)

	if err != nil {
		return "", err
	}

	if isOldBackupName(filepath.Base(resolved)) {
		return "", fmt.Errorf("resolved executable %q is a saved pre-update backup, not the current binary", resolved)
	}

	return resolved, nil
}

// oldBinaryPath is where the pre-update binary is saved during a swap. It is
// the same name selfupdate itself defaults to; the point of computing it
// here and passing it as Options.OldSavePath explicitly is that Apply and
// CleanupOld are then guaranteed to agree on the name.
func oldBinaryPath(target string) string {
	dir := filepath.Dir(target)
	base := filepath.Base(target)

	return filepath.Join(dir, fmt.Sprintf(".%s.old", base))
}

// RollbackFailedError means the swap itself failed and selfupdate's
// automatic rollback to the pre-update binary also failed: there is no file
// at the target path at all. Callers must not suggest retrying — the target
// is gone — and should instead point the user at the saved old binary, if
// it exists.
type RollbackFailedError struct {
	// OldPath is where the pre-update binary would have been saved.
	OldPath string
	Err     error
}

func (e *RollbackFailedError) Error() string {
	return fmt.Sprintf("update failed and the previous binary could not be restored automatically: %v", e.Err)
}

func (e *RollbackFailedError) Unwrap() error {
	return e.Err
}

// PreflightWritable reports whether the update can be installed at all. The
// swap renames a directory entry, so the *directory* must be writable, not the
// binary. Checked before downloading, so a user who cannot install never waits
// for the transfer.
func PreflightWritable() error {
	exe, err := ExecutablePath()

	if err != nil {
		return err
	}

	dir := filepath.Dir(exe)

	probe, err := os.CreateTemp(dir, ".farental-update-*")

	if err != nil {
		return fmt.Errorf("%s: %w", dir, err)
	}

	name := probe.Name()

	probe.Close()

	return os.Remove(name)
}

// progressReader counts bytes as they are read.
type progressReader struct {
	inner   io.Reader
	counter *atomic.Int64
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.inner.Read(b)

	if n > 0 && p.counter != nil {
		p.counter.Add(int64(n))
	}

	return n, err
}

// Apply downloads the release binary and swaps it in. selfupdate.Apply writes
// the new file beside the target, verifies the checksum, performs the rename
// dance for the platform, and rolls back if the swap fails partway.
func Apply(baseURL string, f FileInfo, progress *atomic.Int64) error {
	exe, err := ExecutablePath()

	if err != nil {
		return err
	}

	if f.SizeBytes <= 0 {
		return fmt.Errorf("manifest advertises a non-positive size: %d", f.SizeBytes)
	}

	if f.SizeBytes > maxBinarySize {
		return fmt.Errorf("manifest advertises %d bytes, over the %d byte limit", f.SizeBytes, maxBinarySize)
	}

	if err := requireSecureURL(baseURL); err != nil {
		return err
	}

	checksum, err := hex.DecodeString(f.SHA256)

	if err != nil {
		return fmt.Errorf("invalid checksum in manifest: %w", err)
	}

	endpoint, err := url.JoinPath(strings.TrimSuffix(baseURL, "/"), f.URL)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
	defer cancel()

	// FileDownloadGet sets SetDoNotParseResponse, so resty hands back the
	// live network reader via RawBody() instead of buffering the whole
	// (tens-of-MB) binary into memory first.
	resp, err := request.FileDownloadGet(endpoint).SetContext(ctx).Send()

	if err != nil {
		return err
	}

	// rawBody must be closed on every path below — success, checksum
	// failure, or rollback — or the connection is never released.
	rawBody := resp.RawBody()
	defer rawBody.Close()

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode())
	}

	// The bound is the exact advertised size: a longer body is cut short and
	// fails the checksum, a shorter one fails it too.
	body := &progressReader{
		inner:   io.LimitReader(rawBody, f.SizeBytes),
		counter: progress,
	}

	oldPath := oldBinaryPath(exe)

	err = selfupdate.Apply(body, selfupdate.Options{
		TargetPath: exe,
		Checksum:   checksum,

		// Set explicitly so Apply and CleanupOld always agree on the
		// name: selfupdate's own default happens to match, but leaving
		// it implicit means the two could drift silently in the future.
		// Setting it also means selfupdate never auto-removes the old
		// binary itself; CleanupOld does that at the next startup,
		// uniformly on every platform.
		OldSavePath: oldPath,
	})

	if err != nil {
		// The library documents that callers should always check this:
		// a non-nil rollback error means the second rename failed *and*
		// the rollback rename failed, so there is no file at all at
		// exe. A retry cannot work in that state.
		if rerr := selfupdate.RollbackError(err); rerr != nil {
			return &RollbackFailedError{OldPath: oldPath, Err: err}
		}

		return err
	}

	rememberSwappedPath(exe)

	return nil
}

// CleanupOld removes the previous binary left beside the target by a swap
// that set Options.OldSavePath (see Apply). Harmless if nothing is there.
func CleanupOld() {
	exe, err := ExecutablePath()

	if err != nil {
		return
	}

	os.Remove(oldBinaryPath(exe))
}
