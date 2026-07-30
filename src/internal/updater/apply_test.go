package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func binaryServer(t *testing.T, payload []byte) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(payload)
	}))

	t.Cleanup(srv.Close)

	return srv
}

func sum(payload []byte) string {
	h := sha256.Sum256(payload)
	return hex.EncodeToString(h[:])
}

// applyTo installs into a temp file rather than the test binary itself, and
// resets both package-level globals Apply can mutate so one test's outcome
// never leaks into the next.
func applyTo(t *testing.T, target, baseURL string, f FileInfo, counter *atomic.Int64) error {
	t.Helper()

	originalOverride := targetPathOverride
	t.Cleanup(func() { targetPathOverride = originalOverride })

	targetPathOverride = target

	originalSwapped := swappedPath
	t.Cleanup(func() { swappedPath = originalSwapped })

	return Apply(baseURL, f, counter)
}

func writeFile(t *testing.T, path string, content []byte) {
	t.Helper()

	if err := os.WriteFile(path, content, 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestApplyReplacesTargetAndReportsProgress(t *testing.T) {
	payload := []byte("new binary contents")
	srv := binaryServer(t, payload)

	target := filepath.Join(t.TempDir(), "Farental")
	writeFile(t, target, []byte("old"))

	var counter atomic.Int64

	err := applyTo(t, target, srv.URL, FileInfo{
		Filename:  "Farental",
		SizeBytes: int64(len(payload)),
		SHA256:    sum(payload),
		URL:       "/clienttui/download/42",
	}, &counter)

	if err != nil {
		t.Fatalf("Apply returned %v", err)
	}

	got, err := os.ReadFile(target)

	if err != nil {
		t.Fatal(err)
	}

	if string(got) != string(payload) {
		t.Errorf("target contents = %q, want %q", got, payload)
	}

	if counter.Load() != int64(len(payload)) {
		t.Errorf("progress = %d, want %d", counter.Load(), len(payload))
	}
}

func TestApplyRejectsChecksumMismatch(t *testing.T) {
	srv := binaryServer(t, []byte("tampered"))

	target := filepath.Join(t.TempDir(), "Farental")
	writeFile(t, target, []byte("old"))

	var counter atomic.Int64

	err := applyTo(t, target, srv.URL, FileInfo{
		SizeBytes: int64(len("tampered")),
		SHA256:    sum([]byte("expected something else")),
		URL:       "/clienttui/download/42",
	}, &counter)

	if err == nil {
		t.Fatal("expected a checksum error, got nil")
	}

	got, _ := os.ReadFile(target)

	if string(got) != "old" {
		t.Errorf("target was modified despite the bad checksum: %q", got)
	}
}

// Apply must record the path it swapped into place so Restart can reuse it
// instead of asking the OS again after the swap — see SwappedExecutablePath.
func TestApplyRemembersSwappedPathOnSuccess(t *testing.T) {
	payload := []byte("new binary contents")
	srv := binaryServer(t, payload)

	target := filepath.Join(t.TempDir(), "Farental")
	writeFile(t, target, []byte("old"))

	err := applyTo(t, target, srv.URL, FileInfo{
		SizeBytes: int64(len(payload)),
		SHA256:    sum(payload),
		URL:       "/clienttui/download/42",
	}, new(atomic.Int64))

	if err != nil {
		t.Fatalf("Apply returned %v", err)
	}

	got, err := SwappedExecutablePath()

	if err != nil {
		t.Fatalf("SwappedExecutablePath returned %v", err)
	}

	if got != target {
		t.Errorf("SwappedExecutablePath() = %q, want %q", got, target)
	}
}

// A failed Apply must not make Restart think a swap happened.
func TestApplyDoesNotRememberSwappedPathOnFailure(t *testing.T) {
	srv := binaryServer(t, []byte("tampered"))

	target := filepath.Join(t.TempDir(), "Farental")
	writeFile(t, target, []byte("old"))

	err := applyTo(t, target, srv.URL, FileInfo{
		SizeBytes: int64(len("tampered")),
		SHA256:    sum([]byte("expected something else")),
		URL:       "/clienttui/download/42",
	}, new(atomic.Int64))

	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if swappedPath != "" {
		t.Errorf("swappedPath = %q after a failed Apply, want empty", swappedPath)
	}
}

func TestApplyRejectsZeroSize(t *testing.T) {
	target := filepath.Join(t.TempDir(), "Farental")
	writeFile(t, target, []byte("old"))

	err := applyTo(t, target, "https://example.invalid", FileInfo{
		SizeBytes: 0,
		SHA256:    sum([]byte("x")),
		URL:       "/clienttui/download/1",
	}, new(atomic.Int64))

	if err == nil {
		t.Fatal("expected an error for a zero size, got nil")
	}
}

func TestApplyRejectsNegativeSize(t *testing.T) {
	target := filepath.Join(t.TempDir(), "Farental")
	writeFile(t, target, []byte("old"))

	err := applyTo(t, target, "https://example.invalid", FileInfo{
		SizeBytes: -1,
		SHA256:    sum([]byte("x")),
		URL:       "/clienttui/download/1",
	}, new(atomic.Int64))

	if err == nil {
		t.Fatal("expected an error for a negative size, got nil")
	}
}

func TestApplyRejectsOversizedManifestSize(t *testing.T) {
	target := filepath.Join(t.TempDir(), "Farental")
	writeFile(t, target, []byte("old"))

	err := applyTo(t, target, "https://example.invalid", FileInfo{
		SizeBytes: maxBinarySize + 1,
		SHA256:    sum([]byte("x")),
		URL:       "/clienttui/download/1",
	}, new(atomic.Int64))

	if err == nil {
		t.Fatal("expected an error for a manifest size over the limit, got nil")
	}
}

func TestApplyRejectsNonHTTPSNonLoopbackURL(t *testing.T) {
	target := filepath.Join(t.TempDir(), "Farental")
	writeFile(t, target, []byte("old"))

	err := applyTo(t, target, "http://example.com", FileInfo{
		SizeBytes: 10,
		SHA256:    sum([]byte("x")),
		URL:       "/clienttui/download/1",
	}, new(atomic.Int64))

	if err == nil {
		t.Fatal("expected an error for a non-HTTPS, non-loopback URL, got nil")
	}
}

// Local development serves the manifest and binaries over plain HTTP on
// 127.0.0.1, so that combination must keep working. Every other test in this
// file exercises this path implicitly via httptest.NewServer; this test
// names the requirement directly.
func TestApplyAllowsPlainHTTPOnLoopback(t *testing.T) {
	payload := []byte("new binary contents")
	srv := binaryServer(t, payload)

	target := filepath.Join(t.TempDir(), "Farental")
	writeFile(t, target, []byte("old"))

	err := applyTo(t, target, srv.URL, FileInfo{
		SizeBytes: int64(len(payload)),
		SHA256:    sum(payload),
		URL:       "/clienttui/download/42",
	}, new(atomic.Int64))

	if err != nil {
		t.Fatalf("loopback HTTP should be allowed for local development: %v", err)
	}
}

func TestPreflightWritableRejectsUnwritableDir(t *testing.T) {
	dir := t.TempDir()

	if err := os.Chmod(dir, 0o500); err != nil {
		t.Skip("cannot make the directory read-only here")
	}

	t.Cleanup(func() { os.Chmod(dir, 0o700) })

	original := targetPathOverride
	t.Cleanup(func() { targetPathOverride = original })

	targetPathOverride = filepath.Join(dir, "Farental")

	if err := PreflightWritable(); err == nil {
		t.Error("expected an error for a read-only directory, got nil")
	}
}

func TestPreflightWritableLeavesNoFiles(t *testing.T) {
	dir := t.TempDir()

	original := targetPathOverride
	t.Cleanup(func() { targetPathOverride = original })

	targetPathOverride = filepath.Join(dir, "Farental")

	if err := PreflightWritable(); err != nil {
		t.Fatalf("PreflightWritable returned %v", err)
	}

	entries, err := os.ReadDir(dir)

	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 0 {
		t.Errorf("left %d files behind, want none", len(entries))
	}
}

// CleanupOld must remove exactly the file a real Apply leaves behind, not a
// name this test invents itself: selfupdate's own default backup name is
// ".<basename>.old", not "<basename>.old", and Apply now sets OldSavePath to
// that name explicitly (see oldBinaryPath) so the two never drift apart.
func TestCleanupOldRemovesLeftoverFromARealApply(t *testing.T) {
	payload := []byte("new binary contents")
	srv := binaryServer(t, payload)

	dir := t.TempDir()
	target := filepath.Join(dir, "Farental")
	writeFile(t, target, []byte("old"))

	if err := applyTo(t, target, srv.URL, FileInfo{
		SizeBytes: int64(len(payload)),
		SHA256:    sum(payload),
		URL:       "/clienttui/download/42",
	}, new(atomic.Int64)); err != nil {
		t.Fatalf("Apply returned %v", err)
	}

	// Apply sets OldSavePath, so selfupdate must not have auto-removed the
	// backup itself: it is CleanupOld's job, run at the next startup.
	if _, err := os.Stat(oldBinaryPath(target)); err != nil {
		t.Fatalf("expected a leftover backup after Apply, stat: %v", err)
	}

	CleanupOld()

	entries, err := os.ReadDir(dir)

	if err != nil {
		t.Fatal(err)
	}

	for _, e := range entries {
		if e.Name() != "Farental" {
			t.Errorf("leftover file %q still present after CleanupOld", e.Name())
		}
	}
}

func TestCleanupOldIsHarmlessWithNothingToClean(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "Farental")

	original := targetPathOverride
	t.Cleanup(func() { targetPathOverride = original })

	targetPathOverride = target

	CleanupOld()

	if _, err := os.Stat(oldBinaryPath(target)); !os.IsNotExist(err) {
		t.Errorf("stat = %v, want IsNotExist", err)
	}
}

func TestIsOldBackupName(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{".Farental.old", true},
		{"Farental", false},
		{"Farental.old", false},  // no leading dot: selfupdate never writes this
		{".Farental.new", false}, // the staged file, not the backup
		{".old", true},
	}

	for _, c := range cases {
		if got := isOldBackupName(c.name); got != c.want {
			t.Errorf("isOldBackupName(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestSwappedExecutablePathPrefersRememberedPath(t *testing.T) {
	original := swappedPath
	t.Cleanup(func() { swappedPath = original })

	swappedPath = "/some/remembered/path"

	got, err := SwappedExecutablePath()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "/some/remembered/path" {
		t.Errorf("got %q, want the remembered path", got)
	}
}

func TestSwappedExecutablePathFallsBackWhenNothingRemembered(t *testing.T) {
	originalSwapped := swappedPath
	t.Cleanup(func() { swappedPath = originalSwapped })
	swappedPath = ""

	originalOverride := targetPathOverride
	t.Cleanup(func() { targetPathOverride = originalOverride })
	targetPathOverride = "/tmp/whatever-farental-test-path"

	got, err := SwappedExecutablePath()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "/tmp/whatever-farental-test-path" {
		t.Errorf("got %q, want the override path", got)
	}
}
