package updater

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
)

const sampleManifest = `{
  "version": "1.2.0",
  "published_at": "2026-07-29T10:00:00Z",
  "notes_blocks": [
    {"type": "h2", "spans": [{"text": "Fixes"}]},
    {"type": "p", "spans": [{"text": "Fixed the "}, {"text": "timer", "bold": true}]},
    {"type": "list", "items": [{"indent": 0, "spans": [{"text": "one"}]}]}
  ],
  "files": {
    "PLATFORM": {
      "filename": "Farental",
      "size_bytes": 15499426,
      "sha256": "abc",
      "url": "/clienttui/download/42"
    }
  }
}`

func manifestServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/clienttui/latest" {
			t.Errorf("path = %q, want /clienttui/latest", r.URL.Path)
		}

		if r.URL.Query().Get("lang") != "fr" {
			t.Errorf("lang = %q, want fr", r.URL.Query().Get("lang"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))

	t.Cleanup(srv.Close)

	return srv
}

func TestPlatformKeyMatchesServerConstants(t *testing.T) {
	got := PlatformKey()
	want := runtime.GOOS + "-" + runtime.GOARCH

	if got != want {
		t.Errorf("PlatformKey() = %q, want %q", got, want)
	}

	known := map[string]bool{
		"linux-amd64": true, "linux-arm64": true, "windows-amd64": true,
		"darwin-amd64": true, "darwin-arm64": true,
	}

	if !known[got] {
		t.Logf("running on unsupported platform %q; updater will report no file", got)
	}
}

func TestFetchManifestDecodesBlocks(t *testing.T) {
	body := replacePlatform(sampleManifest)
	srv := manifestServer(t, http.StatusOK, body)

	m, err := fetchManifest(srv.URL, "fr")

	if err != nil {
		t.Fatalf("fetchManifest returned %v", err)
	}

	if m.Version != "1.2.0" {
		t.Errorf("Version = %q, want 1.2.0", m.Version)
	}

	if len(m.NotesBlocks) != 3 {
		t.Fatalf("got %d blocks, want 3", len(m.NotesBlocks))
	}

	if m.NotesBlocks[0].Type != "h2" || m.NotesBlocks[0].Spans[0].Text != "Fixes" {
		t.Errorf("first block = %+v", m.NotesBlocks[0])
	}

	if !m.NotesBlocks[1].Spans[1].Bold {
		t.Errorf("second span should be bold: %+v", m.NotesBlocks[1].Spans[1])
	}

	if m.NotesBlocks[2].Items[0].Spans[0].Text != "one" {
		t.Errorf("list item = %+v", m.NotesBlocks[2].Items)
	}

	f, ok := m.Files[PlatformKey()]

	if !ok {
		t.Fatalf("no entry for %q in %v", PlatformKey(), m.Files)
	}

	if f.SHA256 != "abc" || f.SizeBytes != 15499426 || f.URL != "/clienttui/download/42" {
		t.Errorf("file = %+v", f)
	}
}

func TestFetchManifestNotFound(t *testing.T) {
	srv := manifestServer(t, http.StatusNotFound, `{"error":"no published release"}`)

	if _, err := fetchManifest(srv.URL, "fr"); err == nil {
		t.Fatal("expected an error for 404, got nil")
	}
}

// A stray Taskfile edit or a plain `go build` (config.WebURL's zero value is
// http://127.0.0.1:3001) must not silently make the client fetch and decode
// a manifest, let alone a binary, over cleartext HTTP against a real host.
func TestFetchManifestRejectsNonHTTPSNonLoopbackURL(t *testing.T) {
	if _, err := fetchManifest("http://example.com", "fr"); err == nil {
		t.Fatal("expected an error for a non-HTTPS, non-loopback URL, got nil")
	}
}

// Local development runs the website over plain HTTP on 127.0.0.1
// (manifestServer, via httptest.NewServer, does exactly this), so that
// combination must keep working. Every other test in this file exercises
// it implicitly; this one names the requirement directly.
func TestFetchManifestAllowsPlainHTTPOnLoopback(t *testing.T) {
	body := replacePlatform(sampleManifest)
	srv := manifestServer(t, http.StatusOK, body)

	if _, err := fetchManifest(srv.URL, "fr"); err != nil {
		t.Fatalf("loopback HTTP should be allowed for local development: %v", err)
	}
}

// The JSON body is decoded through a bounded reader, symmetric with the
// binary download's size bound, so a runaway or hostile response body
// cannot make the client buffer an unbounded amount of memory.
func TestFetchManifestRejectsOversizedBody(t *testing.T) {
	huge := strings.Repeat(" ", maxManifestBodySize+1) + replacePlatform(sampleManifest)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(huge))
	}))

	t.Cleanup(srv.Close)

	if _, err := fetchManifest(srv.URL, "fr"); err == nil {
		t.Fatal("expected an error for a body over the size limit, got nil")
	}
}

func TestCheckMandatoryWhenIncompatible(t *testing.T) {
	body := replacePlatform(sampleManifest)
	srv := manifestServer(t, http.StatusOK, body)

	got := checkAt(srv.URL, "1.1.0", "1.2", "fr")

	if got.Mode != ModeMandatory {
		t.Errorf("Mode = %v, want ModeMandatory", got.Mode)
	}

	if got.Latest != "1.2.0" || got.File.SHA256 != "abc" {
		t.Errorf("Result = %+v", got)
	}
}

// An unreachable manifest must not hide the fact that the server refuses this
// client: the mode stays mandatory and the error is carried for the screen.
func TestCheckMandatoryWithUnreachableManifest(t *testing.T) {
	got := checkAt("http://127.0.0.1:1", "1.1.0", "1.2", "fr")

	if got.Mode != ModeMandatory {
		t.Errorf("Mode = %v, want ModeMandatory", got.Mode)
	}

	if got.Err == nil {
		t.Error("Err = nil, want the fetch failure")
	}
}

func TestCheckOptionalWhenNewerPatchExists(t *testing.T) {
	body := replacePlatform(sampleManifest)
	srv := manifestServer(t, http.StatusOK, body)

	got := checkAt(srv.URL, "1.2.0", "1.2", "fr")

	if got.Mode != ModeNone {
		t.Errorf("Mode = %v, want ModeNone for an up-to-date client", got.Mode)
	}

	got = checkAt(srv.URL, "1.1.0", "1.1", "fr")

	if got.Mode != ModeOptional {
		t.Errorf("Mode = %v, want ModeOptional", got.Mode)
	}
}

// A compatible client is never blocked by a failed update check.
func TestCheckCompatibleWithUnreachableManifest(t *testing.T) {
	got := checkAt("http://127.0.0.1:1", "1.1.0", "1.1", "fr")

	if got.Mode != ModeNone {
		t.Errorf("Mode = %v, want ModeNone", got.Mode)
	}
}

func replacePlatform(body string) string {
	return strings.Replace(body, "PLATFORM", PlatformKey(), 1)
}
