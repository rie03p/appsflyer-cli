package appsflyer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultOneLinkBaseURL is the host serving the OneLink API, which is
// separate from the reporting APIs and uses its own token.
const DefaultOneLinkBaseURL = "https://onelink.appsflyer.com"

// OneLinkClient calls the OneLink API v2.0 using a OneLink API token
// (not the API V2 reporting token).
type OneLinkClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type OneLinkOption func(*OneLinkClient)

func WithOneLinkBaseURL(u string) OneLinkOption {
	return func(c *OneLinkClient) { c.baseURL = strings.TrimRight(u, "/") }
}

func WithOneLinkTimeout(d time.Duration) OneLinkOption {
	return func(c *OneLinkClient) { c.httpClient.Timeout = d }
}

func NewOneLink(token string, opts ...OneLinkOption) *OneLinkClient {
	c := &OneLinkClient{
		baseURL:    DefaultOneLinkBaseURL,
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// ShortlinkParams is the request body for creating or updating a
// OneLink short link. Data holds the attribution query parameters and
// must include the media source (pid) on create.
type ShortlinkParams struct {
	ShortlinkID string            `json:"shortlink_id,omitempty"`
	TTL         string            `json:"ttl,omitempty"`
	BrandDomain string            `json:"brand_domain,omitempty"`
	RenewTTL    bool              `json:"renew_ttl,omitempty"`
	Data        map[string]string `json:"data"`
}

// CreateShortlink creates a short link under the given OneLink template
// ID. The caller must close the returned ReadCloser.
func (c *OneLinkClient) CreateShortlink(ctx context.Context, templateID string, p ShortlinkParams) (io.ReadCloser, error) {
	if p.Data["pid"] == "" {
		return nil, fmt.Errorf("appsflyer: the pid (media source) data parameter is required to create a link")
	}
	return c.do(ctx, http.MethodPost, "/api/v2.0/shortlinks/"+url.PathEscape(templateID), p)
}

// GetShortlink fetches a short link's parameters. The caller must close
// the returned ReadCloser.
func (c *OneLinkClient) GetShortlink(ctx context.Context, templateID, shortlinkID string) (io.ReadCloser, error) {
	return c.do(ctx, http.MethodGet, shortlinkPath(templateID, shortlinkID), nil)
}

// UpdateShortlink replaces a short link's parameters. The caller must
// close the returned ReadCloser.
func (c *OneLinkClient) UpdateShortlink(ctx context.Context, templateID, shortlinkID string, p ShortlinkParams) (io.ReadCloser, error) {
	if len(p.Data) == 0 {
		return nil, fmt.Errorf("appsflyer: data parameters are required to update a link")
	}
	return c.do(ctx, http.MethodPut, shortlinkPath(templateID, shortlinkID), p)
}

func (c *OneLinkClient) DeleteShortlink(ctx context.Context, templateID, shortlinkID string) error {
	body, err := c.do(ctx, http.MethodDelete, shortlinkPath(templateID, shortlinkID), nil)
	if err != nil {
		return err
	}
	return body.Close()
}

func shortlinkPath(templateID, shortlinkID string) string {
	return "/api/v2.0/shortlinks/" + url.PathEscape(templateID) + "/" + url.PathEscape(shortlinkID)
}

func (c *OneLinkClient) do(ctx context.Context, method, path string, body any) (io.ReadCloser, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	// The OneLink API expects the raw token, not a Bearer scheme.
	req.Header.Set("Authorization", c.token)
	req.Header.Set("Accept", "application/json")
	if reader != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		resp.Body.Close()
		return nil, &APIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}
	return resp.Body, nil
}
