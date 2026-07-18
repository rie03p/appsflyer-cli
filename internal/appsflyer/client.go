// Package appsflyer provides a client for the AppsFlyer reporting APIs
// (Pull API raw data, Pull API aggregate data, and Master API).
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

// DefaultBaseURL is the host serving the V2-token reporting APIs.
const DefaultBaseURL = "https://hq1.appsflyer.com"

// Client calls the AppsFlyer reporting APIs using an API V2 (Bearer) token.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the API host, mainly for testing.
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") }
}

// WithHTTPClient replaces the underlying HTTP client.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.httpClient = h }
}

// WithTimeout sets the request timeout on the default HTTP client.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.httpClient.Timeout = d }
}

// New returns a Client authenticated with the given API V2 token.
func New(token string, opts ...Option) *Client {
	c := &Client{
		baseURL:    DefaultBaseURL,
		token:      token,
		httpClient: &http.Client{Timeout: 5 * time.Minute},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// APIError is a non-2xx response from the API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	body := strings.TrimSpace(e.Body)
	if len(body) > 500 {
		body = body[:500] + "..."
	}
	msg := fmt.Sprintf("appsflyer: HTTP %d", e.StatusCode)
	switch e.StatusCode {
	case http.StatusUnauthorized:
		msg += " (check that your API V2 token is valid and not revoked)"
	case http.StatusForbidden:
		msg += " (token does not match the app, or the report is not in your subscription)"
	case http.StatusNotFound:
		msg += " (app not found or API not included in your plan)"
	}
	if body != "" {
		msg += ": " + body
	}
	return msg
}

// The caller must close the returned ReadCloser.
func (c *Client) get(ctx context.Context, path string, query url.Values) (io.ReadCloser, error) {
	return c.do(ctx, http.MethodGet, path, query, nil)
}

// The caller must close the returned ReadCloser.
func (c *Client) postJSON(ctx context.Context, path string, query url.Values, body any) (io.ReadCloser, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, http.MethodPost, path, query, bytes.NewReader(b))
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body io.Reader) (io.ReadCloser, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "text/csv, application/json")
	if body != nil {
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
