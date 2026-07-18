package appsflyer

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"
)

// CohortParams is the JSON body for the Cohort API. CohortType, From, To,
// KPIs, AggregationType, and Groupings are required. The lookback window
// is limited to 720 days by the API.
type CohortParams struct {
	CohortType        string              `json:"cohort_type"`
	From              string              `json:"from"`
	To                string              `json:"to"`
	KPIs              []string            `json:"kpis"`
	AggregationType   string              `json:"aggregation_type"`
	Groupings         []string            `json:"groupings"`
	MinCohortSize     int                 `json:"min_cohort_size,omitempty"`
	Granularity       string              `json:"granularity,omitempty"`
	PartialData       bool                `json:"partial_data,omitempty"`
	Filters           map[string][]string `json:"-"`
	PreferredCurrency bool                `json:"preferred_currency,omitempty"`
	PreferredTimezone bool                `json:"preferred_timezone,omitempty"`
	PerUser           bool                `json:"per_user,omitempty"`
	Format            string              `json:"-"` // "" (json) or "csv"
}

func (p CohortParams) validate() error {
	if p.CohortType == "" {
		return fmt.Errorf("appsflyer: cohort_type is required (user_acquisition, retargeting, or unified)")
	}
	if p.From == "" || p.To == "" {
		return fmt.Errorf("appsflyer: from and to dates are required")
	}
	if len(p.KPIs) == 0 {
		return fmt.Errorf("appsflyer: at least one KPI is required")
	}
	if p.AggregationType == "" {
		return fmt.Errorf("appsflyer: aggregation_type is required (cumulative or on_day)")
	}
	if len(p.Groupings) == 0 || len(p.Groupings) > 7 {
		return fmt.Errorf("appsflyer: between 1 and 7 groupings are required")
	}
	return nil
}

// body flattens Filters into the request object, converting the period
// filter to integers as the API expects numeric cohort days.
func (p CohortParams) body() (map[string]any, error) {
	m := map[string]any{
		"cohort_type":      p.CohortType,
		"from":             p.From,
		"to":               p.To,
		"kpis":             p.KPIs,
		"aggregation_type": p.AggregationType,
		"groupings":        p.Groupings,
	}
	if p.MinCohortSize > 0 {
		m["min_cohort_size"] = p.MinCohortSize
	}
	if p.Granularity != "" {
		m["granularity"] = p.Granularity
	}
	if p.PartialData {
		m["partial_data"] = true
	}
	if p.PreferredCurrency {
		m["preferred_currency"] = true
	}
	if p.PreferredTimezone {
		m["preferred_timezone"] = true
	}
	if p.PerUser {
		m["per_user"] = true
	}
	if len(p.Filters) > 0 {
		filters := make(map[string]any, len(p.Filters))
		for key, vals := range p.Filters {
			if key == "period" {
				periods := make([]int, len(vals))
				for i, v := range vals {
					n, err := strconv.Atoi(v)
					if err != nil {
						return nil, fmt.Errorf("appsflyer: period filter values must be integers, got %q", v)
					}
					periods[i] = n
				}
				filters[key] = periods
				continue
			}
			filters[key] = vals
		}
		m["filters"] = filters
	}
	return m, nil
}

// Cohort streams a Cohort API report. The caller must close the
// returned ReadCloser.
func (c *Client) Cohort(ctx context.Context, appID string, p CohortParams) (io.ReadCloser, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}
	body, err := p.body()
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	setOpt(q, "format", p.Format)
	path := "/api/cohorts/v1/data/app/" + url.PathEscape(appID)
	return c.postJSON(ctx, path, q, body)
}
