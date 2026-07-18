package cli

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func isolateConfig(t *testing.T) {
	t.Helper()
	t.Setenv("AFCLI_CONFIG_DIR", t.TempDir())
	t.Setenv("APPSFLYER_API_TOKEN", "")
}

func TestAuthLoginStoresTokenUsedByCommands(t *testing.T) {
	isolateConfig(t)

	out, err := run(t, "auth", "login", "--token", "tok")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Token saved to") {
		t.Errorf("out = %q", out)
	}

	srv := newServer(t, "/api/raw-data/export/app/id1/installs_report/v5", "csv")
	out, err = run(t,
		"raw", "installs_report",
		"--app", "id1", "--from", "2026-07-01", "--to", "2026-07-07",
		"--base-url", srv.URL,
	)
	if err != nil {
		t.Fatal(err)
	}
	if out != "csv" {
		t.Errorf("out = %q", out)
	}
}

func TestAuthLoginPromptsWhenNoFlag(t *testing.T) {
	isolateConfig(t)

	cmd := NewRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	cmd.SetIn(strings.NewReader("piped-token\n"))
	cmd.SetArgs([]string{"auth", "login"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	status, err := run(t, "auth", "status")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(status, "piped-to...") {
		t.Errorf("status = %q", status)
	}
}

func TestAuthStatus(t *testing.T) {
	isolateConfig(t)

	out, err := run(t, "auth", "status")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Not logged in") {
		t.Errorf("out = %q", out)
	}

	t.Setenv("APPSFLYER_API_TOKEN", "env-token-value")
	out, err = run(t, "auth", "status")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "APPSFLYER_API_TOKEN") || !strings.Contains(out, "env-toke...") {
		t.Errorf("out = %q", out)
	}
}

func TestAuthLogout(t *testing.T) {
	isolateConfig(t)

	if _, err := run(t, "auth", "login", "--token", "tok"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "auth", "logout")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "deleted") {
		t.Errorf("out = %q", out)
	}

	status, _ := run(t, "auth", "status")
	if !strings.Contains(status, "Not logged in") {
		t.Errorf("status after logout = %q", status)
	}
}
