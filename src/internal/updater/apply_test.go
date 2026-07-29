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

// applyTo installs into a temp file rather than the test binary itself.
func applyTo(t *testing.T, target, baseURL string, f FileInfo, counter *atomic.Int64) error {
	t.Helper()

	original := targetPathOverride
	t.Cleanup(func() { targetPathOverride = original })

	targetPathOverride = target

	return Apply(baseURL, f, counter)
}

func TestApplyReplacesTargetAndReportsProgress(t *testing.T) {
	payload := []byte("new binary contents")
	srv := binaryServer(t, payload)

	target := filepath.Join(t.TempDir(), "Farental")

	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

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

	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

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

func TestCleanupOldRemovesLeftover(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "Farental")

	if err := os.WriteFile(target+".old", []byte("stale"), 0o755); err != nil {
		t.Fatal(err)
	}

	original := targetPathOverride
	t.Cleanup(func() { targetPathOverride = original })

	targetPathOverride = target

	CleanupOld()

	if _, err := os.Stat(target + ".old"); !os.IsNotExist(err) {
		t.Errorf("leftover .old still present (err = %v)", err)
	}
}
