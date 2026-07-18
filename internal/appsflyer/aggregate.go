package appsflyer

import (
	"context"
	"fmt"
	"io"
	"net/url"
)

// AggReports lists the report names accepted by the aggregate data Pull API.
var AggReports = map[string]string{
	"partners_report":         "LTV data grouped by media source and campaign",
	"partners_by_date_report": "LTV data grouped by date, media source, and campaign",
	"daily_report":            "LTV data grouped by date, media source, and campaign (no in-app events)",
	"geo_report":              "LTV data grouped by geo, media source, and campaign",
	"geo_by_date_report":      "LTV data grouped by date, geo, media source, and campaign",
}

// AggregateParams are the query parameters for aggregate data reports.
// From and To are required, in "yyyy-mm-dd" format.
type AggregateParams struct {
	From                 string
	To                   string
	MediaSource          string
	Category             string
	AttributionTouchType string
	Currency             string
	Timezone             string
	Reattr               bool
	Format               string // "" (CSV) or "json"
}

func (p AggregateParams) values() (url.Values, error) {
	if p.From == "" || p.To == "" {
		return nil, fmt.Errorf("appsflyer: from and to dates are required")
	}
	v := url.Values{}
	v.Set("from", p.From)
	v.Set("to", p.To)
	setOpt(v, "media_source", p.MediaSource)
	setOpt(v, "category", p.Category)
	setOpt(v, "attribution_touch_type", p.AttributionTouchType)
	setOpt(v, "currency", p.Currency)
	setOpt(v, "timezone", p.Timezone)
	if p.Reattr {
		v.Set("reattr", "true")
	}
	setOpt(v, "format", p.Format)
	return v, nil
}

// Aggregate streams an aggregate data report. The caller must close the
// returned ReadCloser.
func (c *Client) Aggregate(ctx context.Context, appID, report string, p AggregateParams) (io.ReadCloser, error) {
	v, err := p.values()
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/api/agg-data/export/app/%s/%s/v5", url.PathEscape(appID), url.PathEscape(report))
	return c.get(ctx, path, v)
}
