package appsflyer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestCohort(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q", r.Method)
		}
		if r.URL.Path != "/api/cohorts/v1/data/app/id123" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("format"); got != "csv" {
			t.Errorf("format = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["cohort_type"] != "user_acquisition" {
			t.Errorf("cohort_type = %v", body["cohort_type"])
		}
		if body["aggregation_type"] != "cumulative" {
			t.Errorf("aggregation_type = %v", body["aggregation_type"])
		}
		filters, _ := body["filters"].(map[string]any)
		periods, _ := filters["period"].([]any)
		if len(periods) != 2 || periods[0] != float64(0) || periods[1] != float64(7) {
			t.Errorf("filters.period = %v, want [0 7] as numbers", filters["period"])
		}
		pids, _ := filters["pid"].([]any)
		if len(pids) != 1 || pids[0] != "facebook" {
			t.Errorf("filters.pid = %v", filters["pid"])
		}
		if _, present := body["min_cohort_size"]; present {
			t.Error("min_cohort_size should be omitted when zero")
		}
		io.WriteString(w, "cohort-csv")
	})

	body, err := c.Cohort(context.Background(), "id123", CohortParams{
		CohortType:      "user_acquisition",
		From:            "2026-06-01",
		To:              "2026-06-30",
		KPIs:            []string{"users", "roas"},
		AggregationType: "cumulative",
		Groupings:       []string{"pid", "geo"},
		Filters:         map[string][]string{"period": {"0", "7"}, "pid": {"facebook"}},
		Format:          "csv",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := readAll(t, body); got != "cohort-csv" {
		t.Errorf("body = %q", got)
	}
}

func TestCohortValidation(t *testing.T) {
	c := New("t")
	base := CohortParams{
		CohortType:      "unified",
		From:            "2026-06-01",
		To:              "2026-06-30",
		KPIs:            []string{"users"},
		AggregationType: "on_day",
		Groupings:       []string{"pid"},
	}

	for name, mutate := range map[string]func(*CohortParams){
		"missing cohort_type":      func(p *CohortParams) { p.CohortType = "" },
		"missing dates":            func(p *CohortParams) { p.From = "" },
		"missing kpis":             func(p *CohortParams) { p.KPIs = nil },
		"missing aggregation_type": func(p *CohortParams) { p.AggregationType = "" },
		"too many groupings": func(p *CohortParams) {
			p.Groupings = []string{"a", "b", "c", "d", "e", "f", "g", "h"}
		},
		"non-numeric period": func(p *CohortParams) {
			p.Filters = map[string][]string{"period": {"abc"}}
		},
	} {
		p := base
		mutate(&p)
		if _, err := c.Cohort(context.Background(), "id", p); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}
