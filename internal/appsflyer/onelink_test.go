package appsflyer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestOneLinkClient(t *testing.T, handler http.HandlerFunc) *OneLinkClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewOneLink("ol-token", WithOneLinkBaseURL(srv.URL))
}

func TestCreateShortlink(t *testing.T) {
	c := newTestOneLinkClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q", r.Method)
		}
		if r.URL.Path != "/api/v2.0/shortlinks/abc123" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "ol-token" {
			t.Errorf("Authorization = %q, want the raw token without a Bearer prefix", got)
		}
		var p ShortlinkParams
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Fatal(err)
		}
		if p.Data["pid"] != "email" || p.TTL != "90d" || p.ShortlinkID != "promo" {
			t.Errorf("body = %+v", p)
		}
		io.WriteString(w, `{"url":"https://myapp.onelink.me/abc123/promo"}`)
	})

	body, err := c.CreateShortlink(context.Background(), "abc123", ShortlinkParams{
		ShortlinkID: "promo",
		TTL:         "90d",
		Data:        map[string]string{"pid": "email", "c": "summer"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := readAll(t, body); got == "" {
		t.Error("empty body")
	}
}

func TestCreateShortlinkRequiresPid(t *testing.T) {
	c := NewOneLink("t")
	_, err := c.CreateShortlink(context.Background(), "abc123", ShortlinkParams{
		Data: map[string]string{"c": "summer"},
	})
	if err == nil {
		t.Fatal("expected error for missing pid")
	}
}

func TestGetUpdateDeleteShortlink(t *testing.T) {
	var gotMethods []string
	c := newTestOneLinkClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethods = append(gotMethods, r.Method)
		if r.URL.Path != "/api/v2.0/shortlinks/abc123/qwer9876" {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, "{}")
	})

	ctx := context.Background()
	if body, err := c.GetShortlink(ctx, "abc123", "qwer9876"); err != nil {
		t.Fatal(err)
	} else {
		body.Close()
	}
	if body, err := c.UpdateShortlink(ctx, "abc123", "qwer9876", ShortlinkParams{
		Data: map[string]string{"pid": "sms"},
	}); err != nil {
		t.Fatal(err)
	} else {
		body.Close()
	}
	if err := c.DeleteShortlink(ctx, "abc123", "qwer9876"); err != nil {
		t.Fatal(err)
	}

	want := []string{http.MethodGet, http.MethodPut, http.MethodDelete}
	if len(gotMethods) != 3 || gotMethods[0] != want[0] || gotMethods[1] != want[1] || gotMethods[2] != want[2] {
		t.Errorf("methods = %v, want %v", gotMethods, want)
	}
}

func TestUpdateShortlinkRequiresData(t *testing.T) {
	c := NewOneLink("t")
	if _, err := c.UpdateShortlink(context.Background(), "a", "b", ShortlinkParams{}); err == nil {
		t.Fatal("expected error for empty data")
	}
}
