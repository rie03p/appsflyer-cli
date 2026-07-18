package appsflyer

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New("test-token", WithBaseURL(srv.URL))
}

func readAll(t *testing.T, r io.ReadCloser) string {
	t.Helper()
	defer r.Close()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestRawData(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		if r.URL.Path != "/api/raw-data/export/app/id123/installs_report/v5" {
			t.Errorf("path = %q", r.URL.Path)
		}
		q := r.URL.Query()
		for key, want := range map[string]string{
			"from":         "2026-07-01",
			"to":           "2026-07-07",
			"media_source": "facebook",
			"event_name":   "af_purchase,af_login",
			"maximum_rows": "100",
		} {
			if got := q.Get(key); got != want {
				t.Errorf("query %s = %q, want %q", key, got, want)
			}
		}
		io.WriteString(w, "csv-data")
	})

	body, err := c.RawData(context.Background(), "id123", "installs_report", RawDataParams{
		From:        "2026-07-01",
		To:          "2026-07-07",
		MediaSource: "facebook",
		EventName:   []string{"af_purchase", "af_login"},
		MaximumRows: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := readAll(t, body); got != "csv-data" {
		t.Errorf("body = %q", got)
	}
}

func TestRawDataRequiresDates(t *testing.T) {
	c := New("t")
	if _, err := c.RawData(context.Background(), "id", "installs_report", RawDataParams{}); err == nil {
		t.Fatal("expected error for missing dates")
	}
}

func TestAggregate(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/agg-data/export/app/id123/partners_report/v5" {
			t.Errorf("path = %q", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("reattr") != "true" {
			t.Errorf("reattr = %q", q.Get("reattr"))
		}
		if q.Get("format") != "json" {
			t.Errorf("format = %q", q.Get("format"))
		}
		io.WriteString(w, "agg")
	})

	body, err := c.Aggregate(context.Background(), "id123", "partners_report", AggregateParams{
		From:   "2026-07-01",
		To:     "2026-07-07",
		Reattr: true,
		Format: "json",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := readAll(t, body); got != "agg" {
		t.Errorf("body = %q", got)
	}
}

func TestMaster(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/master-agg-data/v4/app/id123" {
			t.Errorf("path = %q", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("groupings") != "pid,geo" {
			t.Errorf("groupings = %q", q.Get("groupings"))
		}
		if q.Get("kpis") != "installs,clicks" {
			t.Errorf("kpis = %q", q.Get("kpis"))
		}
		if q.Get("pid") != "facebook" {
			t.Errorf("pid = %q", q.Get("pid"))
		}
		if q.Get("calculated_kpi_ctr") != "clicks/impressions" {
			t.Errorf("calculated_kpi_ctr = %q", q.Get("calculated_kpi_ctr"))
		}
		io.WriteString(w, "master")
	})

	body, err := c.Master(context.Background(), "id123", MasterParams{
		From:           "2026-07-01",
		To:             "2026-07-07",
		Groupings:      []string{"pid", "geo"},
		KPIs:           []string{"installs", "clicks"},
		Filters:        map[string]string{"pid": "facebook"},
		CalculatedKPIs: map[string]string{"ctr": "clicks/impressions"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := readAll(t, body); got != "master" {
		t.Errorf("body = %q", got)
	}
}

func TestMasterRequiresGroupingsAndKPIs(t *testing.T) {
	c := New("t")
	_, err := c.Master(context.Background(), "id", MasterParams{From: "2026-07-01", To: "2026-07-02"})
	if err == nil {
		t.Fatal("expected error for missing groupings")
	}
}

func TestAPIError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "bad token")
	})

	_, err := c.Aggregate(context.Background(), "id", "daily_report", AggregateParams{
		From: "2026-07-01", To: "2026-07-07",
	})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %v", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d", apiErr.StatusCode)
	}
}
