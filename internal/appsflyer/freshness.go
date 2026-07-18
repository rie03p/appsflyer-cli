package appsflyer

import (
	"context"
	"io"
)

// Freshness streams the Master API data freshness report (when the
// aggregated data was last updated). The caller must close the
// returned ReadCloser.
func (c *Client) Freshness(ctx context.Context) (io.ReadCloser, error) {
	return c.get(ctx, "/api/master-agg-data/lastupdate", nil)
}
