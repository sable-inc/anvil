// Package hyperdx provides an HTTP client for the HyperDX observability API.
package hyperdx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

// Doer executes HTTP requests. *http.Client satisfies this.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// APIError represents an HTTP error response from the HyperDX API.
type APIError struct {
	StatusCode int
	Method     string
	Path       string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("HyperDX API %s %s returned %d: %s", e.Method, e.Path, e.StatusCode, e.Body)
}

// IsUnauthorized returns true if the error is a 401 response.
func IsUnauthorized(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401
	}
	return false
}

// Client communicates with the HyperDX API.
type Client struct {
	baseURL string
	doer    Doer
}

// NewClient creates a HyperDX API client with Bearer token auth.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		doer: &http.Client{
			Timeout:   defaultTimeout,
			Transport: &bearerTransport{apiKey: apiKey},
		},
	}
}

// NewClientWithDoer creates a client with a custom Doer (for testing).
func NewClientWithDoer(baseURL string, doer Doer) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), doer: doer}
}

// Get performs a GET request and returns the raw JSON response.
func (c *Client) Get(ctx context.Context, path string) (json.RawMessage, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

// Post performs a POST request with a JSON body and returns the raw JSON response.
func (c *Client) Post(ctx context.Context, path string, body any) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPost, path, body)
}

func (c *Client) do(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.doer.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Method:     method,
			Path:       path,
			Body:       string(respBody),
		}
	}

	return json.RawMessage(respBody), nil
}

// bearerTransport injects Authorization: Bearer <apiKey> on every request.
type bearerTransport struct {
	apiKey string
	base   http.RoundTripper
}

func (t *bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	if t.apiKey != "" {
		req2.Header.Set("Authorization", "Bearer "+t.apiKey)
	}
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req2)
}
