package cli

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// run executes the command tree with args and returns stdout.
func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := NewRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func newServer(t *testing.T, wantPath string, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("Authorization = %q", got)
		}
		io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestRawCommand(t *testing.T) {
	srv := newServer(t, "/api/raw-data/export/app/id1/installs_report/v5", "raw-csv")
	out, err := run(t,
		"raw", "installs_report",
		"--app", "id1", "--from", "2026-07-01", "--to", "2026-07-07",
		"--token", "tok", "--base-url", srv.URL,
	)
	if err != nil {
		t.Fatal(err)
	}
	if out != "raw-csv" {
		t.Errorf("out = %q", out)
	}
}

func TestAggCommandWritesFile(t *testing.T) {
	srv := newServer(t, "/api/agg-data/export/app/id1/daily_report/v5", "agg-csv")
	dest := filepath.Join(t.TempDir(), "out.csv")
	_, err := run(t,
		"agg", "daily_report",
		"--app", "id1", "--from", "2026-07-01", "--to", "2026-07-07",
		"--token", "tok", "--base-url", srv.URL, "-o", dest,
	)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "agg-csv" {
		t.Errorf("file = %q", b)
	}
}

func TestMasterCommandRejectsUnknownFilter(t *testing.T) {
	_, err := run(t,
		"master",
		"--app", "id1", "--from", "2026-07-01", "--to", "2026-07-07",
		"--groupings", "pid", "--kpis", "installs",
		"--filter", "bogus=x", "--token", "tok",
	)
	if err == nil || !strings.Contains(err.Error(), "unknown filter") {
		t.Fatalf("err = %v, want unknown filter error", err)
	}
}

func TestMissingTokenError(t *testing.T) {
	t.Setenv("AFCLI_CONFIG_DIR", t.TempDir())
	t.Setenv("APPSFLYER_API_TOKEN", "")
	_, err := run(t,
		"raw", "installs_report",
		"--app", "id1", "--from", "2026-07-01", "--to", "2026-07-07",
	)
	if err == nil || !strings.Contains(err.Error(), "no API token") {
		t.Fatalf("err = %v, want token error", err)
	}
}
