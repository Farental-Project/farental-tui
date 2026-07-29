package updater

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/minio/selfupdate"
)

// downloadTimeout bounds the whole transfer, not just the connection.
const downloadTimeout = 10 * time.Minute

// targetPathOverride replaces the running executable's path in tests. Empty in
// production, where os.Executable is used.
var targetPathOverride string

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

	return filepath.EvalSymlinks(exe)
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

	checksum, err := hex.DecodeString(f.SHA256)

	if err != nil {
		return fmt.Errorf("invalid checksum in manifest: %w", err)
	}

	endpoint, err := url.JoinPath(strings.TrimSuffix(baseURL, "/"), f.URL)

	if err != nil {
		return err
	}

	client := &http.Client{Timeout: downloadTimeout}

	resp, err := client.Get(endpoint)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// The bound is the exact advertised size: a longer body is cut short and
	// fails the checksum, a shorter one fails it too.
	body := &progressReader{
		inner:   io.LimitReader(resp.Body, f.SizeBytes),
		counter: progress,
	}

	return selfupdate.Apply(body, selfupdate.Options{
		TargetPath: exe,
		Checksum:   checksum,
	})
}

// CleanupOld removes the previous binary left beside the target. Windows
// cannot delete a running executable, so the swap renames it aside and the
// next launch clears it. Harmless elsewhere.
func CleanupOld() {
	exe, err := ExecutablePath()

	if err != nil {
		return
	}

	os.Remove(exe + ".old")
}
