package api

import (
	"bytes"
	"context"
	"encoding/json"
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

// Client communicates with the Sable API.
type Client struct {
	baseURL string
	doer    Doer
}

// NewClient creates an API client for the given base URL and auth token.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		doer: &http.Client{
			Timeout:   defaultTimeout,
			Transport: &AuthTransport{Token: token},
		},
	}
}

// NewClientWithDoer creates a client with a custom Doer (for testing).
func NewClientWithDoer(baseURL string, doer Doer) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), doer: doer}
}

// Get performs a GET request and decodes the JSON response into v.
func (c *Client) Get(ctx context.Context, path string, v any) error {
	return c.do(ctx, http.MethodGet, path, nil, v)
}

// Post performs a POST request with a JSON body and decodes the response into v.
func (c *Client) Post(ctx context.Context, path string, body, v any) error {
	return c.do(ctx, http.MethodPost, path, body, v)
}

// Put performs a PUT request with a JSON body and decodes the response into v.
func (c *Client) Put(ctx context.Context, path string, body, v any) error {
	return c.do(ctx, http.MethodPut, path, body, v)
}

// Patch performs a PATCH request with a JSON body and decodes the response into v.
func (c *Client) Patch(ctx context.Context, path string, body, v any) error {
	return c.do(ctx, http.MethodPatch, path, body, v)
}

// Delete performs a DELETE request and decodes the response into v.
func (c *Client) Delete(ctx context.Context, path string, v any) error {
	return c.do(ctx, http.MethodDelete, path, nil, v)
}

// do executes an HTTP request, handles errors, and decodes the response.
func (c *Client) do(ctx context.Context, method, path string, body, v any) error {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.doer.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	// io.ReadAll uses 50% less memory in Go 1.26.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return parseErrorResponse(resp.StatusCode, respBody)
	}

	if v != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, v); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}

// parseErrorResponse creates a typed error from an error response body.
// sable-api returns { "error": "message", "hint": "optional hint" }.
func parseErrorResponse(statusCode int, body []byte) error {
	var errResp struct {
		Error string `json:"error"`
		Hint  string `json:"hint"`
	}
	if err := json.Unmarshal(body, &errResp); err != nil {
		// Couldn't parse — use the raw body as the message.
		return NewFromStatus(statusCode, string(body), "")
	}
	msg := errResp.Error
	if msg == "" {
		msg = http.StatusText(statusCode)
	}
	return NewFromStatus(statusCode, msg, errResp.Hint)
}
