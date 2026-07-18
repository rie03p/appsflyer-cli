package appsflyer

import (
	"context"
	"fmt"
	"io"
	"net/url"
)

// SKANParams are the query parameters for the SKAN aggregated
// performance report. StartDate and EndDate are required; the range is
// limited to 90 days by the API.
type SKANParams struct {
	StartDate string
	EndDate   string
	Version   string // "v3" (SKAN 4 postbacks, default) or "v2" (SKAN 3)
	DateType  string // "install" (default) or "arrival"
	ViewType  string // "unified" (default), "acquisition", or "retargeting"
	Modeled   bool   // modeled conversion values; v2 only
}

// SKAN streams a SKAN aggregated performance report. The caller must
// close the returned ReadCloser.
func (c *Client) SKAN(ctx context.Context, appID string, p SKANParams) (io.ReadCloser, error) {
	if p.StartDate == "" || p.EndDate == "" {
		return nil, fmt.Errorf("appsflyer: start and end dates are required")
	}
	version := p.Version
	if version == "" {
		version = "v3"
	}
	if version != "v2" && version != "v3" {
		return nil, fmt.Errorf("appsflyer: version must be v2 (SKAN 3) or v3 (SKAN 4), got %q", version)
	}
	if p.Modeled && version != "v2" {
		return nil, fmt.Errorf("appsflyer: modeled conversion values are only supported by v2")
	}
	q := url.Values{}
	q.Set("start_date", p.StartDate)
	q.Set("end_date", p.EndDate)
	setOpt(q, "date_type", p.DateType)
	setOpt(q, "view_type", p.ViewType)
	if p.Modeled {
		q.Set("modeled_conversion_values", "true")
	}
	path := fmt.Sprintf("/api/skadnetworks/%s/data/app/%s", version, url.PathEscape(appID))
	return c.get(ctx, path, q)
}
