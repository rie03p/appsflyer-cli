package appsflyer

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestSKAN(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/skadnetworks/v3/data/app/id123" {
			t.Errorf("path = %q", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("start_date") != "2026-07-01" || q.Get("end_date") != "2026-07-07" {
			t.Errorf("dates = %q..%q", q.Get("start_date"), q.Get("end_date"))
		}
		if q.Get("view_type") != "acquisition" {
			t.Errorf("view_type = %q", q.Get("view_type"))
		}
		io.WriteString(w, "skan")
	})

	body, err := c.SKAN(context.Background(), "id123", SKANParams{
		StartDate: "2026-07-01",
		EndDate:   "2026-07-07",
		ViewType:  "acquisition",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := readAll(t, body); got != "skan" {
		t.Errorf("body = %q", got)
	}
}

func TestSKANValidation(t *testing.T) {
	c := New("t")
	cases := map[string]SKANParams{
		"missing dates":   {},
		"bad version":     {StartDate: "2026-07-01", EndDate: "2026-07-07", Version: "v9"},
		"modeled with v3": {StartDate: "2026-07-01", EndDate: "2026-07-07", Modeled: true},
	}
	for name, p := range cases {
		if _, err := c.SKAN(context.Background(), "id", p); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}
