package api

import (
	"net/http"
	"strings"
)

const serviceTokenPrefix = "svc_"

// AuthTransport is an http.RoundTripper that injects the Authorization header.
type AuthTransport struct {
	Token string
	Base  http.RoundTripper
}

// RoundTrip adds the auth header and delegates to the base transport.
func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone to avoid mutating the caller's request.
	req2 := req.Clone(req.Context())
	if t.Token != "" {
		token := t.Token
		if !strings.HasPrefix(token, serviceTokenPrefix) {
			token = serviceTokenPrefix + token
		}
		req2.Header.Set("Authorization", "Bearer "+token)
	}
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req2)
}
