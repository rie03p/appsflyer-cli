package appsflyer

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// MasterFilterKeys are the filter dimensions accepted by the Master API.
var MasterFilterKeys = []string{"pid", "c", "af_prt", "af_channel", "af_siteid", "geo"}

// MasterParams are the query parameters for the Master API.
// From, To, Groupings, and KPIs are required. The date range is limited
// to 31 days by the API.
type MasterParams struct {
	From           string
	To             string
	Groupings      []string
	KPIs           []string
	CalculatedKPIs map[string]string // name (without calculated_kpi_ prefix) -> formula
	Filters        map[string]string // see MasterFilterKeys
	Currency       string
	Timezone       string
	Format         string // "" (CSV) or "json"
}

func (p MasterParams) values() (url.Values, error) {
	if p.From == "" || p.To == "" {
		return nil, fmt.Errorf("appsflyer: from and to dates are required")
	}
	if len(p.Groupings) == 0 {
		return nil, fmt.Errorf("appsflyer: at least one grouping is required")
	}
	if len(p.KPIs) == 0 {
		return nil, fmt.Errorf("appsflyer: at least one KPI is required")
	}
	v := url.Values{}
	v.Set("from", p.From)
	v.Set("to", p.To)
	v.Set("groupings", strings.Join(p.Groupings, ","))
	v.Set("kpis", strings.Join(p.KPIs, ","))
	for name, formula := range p.CalculatedKPIs {
		if !strings.HasPrefix(name, "calculated_kpi_") {
			name = "calculated_kpi_" + name
		}
		v.Set(name, formula)
	}
	for key, val := range p.Filters {
		setOpt(v, key, val)
	}
	setOpt(v, "currency", p.Currency)
	setOpt(v, "timezone", p.Timezone)
	setOpt(v, "format", p.Format)
	return v, nil
}

// Master streams a Master API report for the given app ID (or a
// comma-separated list of app IDs, or "all"). The caller must close the
// returned ReadCloser.
func (c *Client) Master(ctx context.Context, appID string, p MasterParams) (io.ReadCloser, error) {
	v, err := p.values()
	if err != nil {
		return nil, err
	}
	path := "/api/master-agg-data/v4/app/" + url.PathEscape(appID)
	return c.get(ctx, path, v)
}
