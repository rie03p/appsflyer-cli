package appsflyer

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
)

// RawReports lists the report names accepted by the raw data Pull API,
// grouped for help output.
var RawReports = map[string][]string{
	"User acquisition": {
		"installs_report",
		"in_app_events_report",
		"uninstall_events_report",
		"reinstalls",
	},
	"Organic": {
		"organic_installs_report",
		"organic_in_app_events_report",
		"organic_uninstall_events_report",
		"reinstalls_organic",
	},
	"Retargeting": {
		"installs-retarget",
		"in-app-events-retarget",
	},
	"Ad revenue": {
		"ad_revenue_raw",
		"ad_revenue_organic_raw",
		"ad-revenue-raw-retarget",
	},
	"Protect360 fraud": {
		"blocked_installs_report",
		"blocked_in_app_events_report",
		"detection",
		"fraud-post-inapps",
		"blocked_clicks_report",
		"blocked_install_postbacks",
	},
	"Postbacks": {
		"installs-postbacks",
		"in-app-events-postbacks",
		"retarget_install_postbacks",
		"retarget_in_app_events_postbacks",
	},
}

// RawDataParams are the query parameters for raw data reports.
// From and To are required; dates use "yyyy-mm-dd" and may include a
// time component ("yyyy-mm-dd hh:mm:ss").
type RawDataParams struct {
	From             string
	To               string
	MediaSource      string
	Category         string
	EventName        []string
	Timezone         string
	Geo              string
	Currency         string
	MaximumRows      int
	AdditionalFields []string
}

func (p RawDataParams) values() (url.Values, error) {
	if p.From == "" || p.To == "" {
		return nil, fmt.Errorf("appsflyer: from and to dates are required")
	}
	v := url.Values{}
	v.Set("from", p.From)
	v.Set("to", p.To)
	setOpt(v, "media_source", p.MediaSource)
	setOpt(v, "category", p.Category)
	setOpt(v, "event_name", strings.Join(p.EventName, ","))
	setOpt(v, "timezone", p.Timezone)
	setOpt(v, "geo", p.Geo)
	setOpt(v, "currency", p.Currency)
	if p.MaximumRows > 0 {
		v.Set("maximum_rows", strconv.Itoa(p.MaximumRows))
	}
	setOpt(v, "additional_fields", strings.Join(p.AdditionalFields, ","))
	return v, nil
}

// RawData streams a raw data report as CSV. The caller must close the
// returned ReadCloser.
func (c *Client) RawData(ctx context.Context, appID, report string, p RawDataParams) (io.ReadCloser, error) {
	v, err := p.values()
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/api/raw-data/export/app/%s/%s/v5", url.PathEscape(appID), url.PathEscape(report))
	return c.get(ctx, path, v)
}

func setOpt(v url.Values, key, val string) {
	if val != "" {
		v.Set(key, val)
	}
}
