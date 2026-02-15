// Package app provides the dependency container for the Anvil CLI.
package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sable-inc/anvil/internal/api"
)

// App is the top-level dependency container passed to all commands.
// It holds configuration, I/O writers, and shared state.
type App struct {
	Out     io.Writer
	ErrOut  io.Writer
	Format  string
	APIURL  string
	OrgID   string
	Token   string
	Verbose bool
	NoColor bool

	client *api.Client
}

// Option configures an App during construction.
type Option func(*App)

// New creates an App with sensible defaults and applies any options.
func New(opts ...Option) *App {
	a := &App{
		Out:    os.Stdout,
		ErrOut: os.Stderr,
		Format: "table",
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Client returns the lazily-created API client.
// Returns nil if no API URL is configured.
func (a *App) Client() *api.Client {
	if a.client == nil && a.APIURL != "" {
		a.client = api.NewClient(a.APIURL, a.Token)
	}
	return a.client
}

// RequireClient returns the API client or an error if no API URL is configured.
// Use for endpoints that don't require authentication (health, connection-details).
func (a *App) RequireClient() (*api.Client, error) {
	c := a.Client()
	if c == nil {
		return nil, fmt.Errorf("no API URL configured — set --api-url or api_url in config")
	}
	return c, nil
}

// RequireAuth returns the API client or an error if not authenticated.
func (a *App) RequireAuth() (*api.Client, error) {
	if a.Token == "" {
		return nil, fmt.Errorf("not authenticated — run 'anvil auth login' first")
	}
	return a.RequireClient()
}

// RequirePublicID returns the org public ID (org_xxx format) or an error.
// Deploy and LiveKit endpoints use publicId as a path parameter.
// Accepts either a publicId directly (org_xxx) or a numeric orgId (resolves via API).
func (a *App) RequirePublicID(ctx context.Context) (string, error) {
	if a.OrgID == "" {
		return "", fmt.Errorf("--org is required (use org public ID, e.g. org_xxx)")
	}
	if strings.HasPrefix(a.OrgID, "org_") {
		return a.OrgID, nil
	}
	// Numeric orgId — look up the publicId via the organizations endpoint.
	client, err := a.RequireAuth()
	if err != nil {
		return "", err
	}
	var resp struct {
		Organization struct {
			PublicID string `json:"publicId"`
		} `json:"organization"`
	}
	if err := client.Get(ctx, "/organizations/"+a.OrgID, &resp); err != nil {
		return "", fmt.Errorf("resolving org publicId: %w", err)
	}
	if resp.Organization.PublicID == "" {
		return "", fmt.Errorf("organization %s has no publicId", a.OrgID)
	}
	return resp.Organization.PublicID, nil
}

// WithOutput sets the standard output writer.
func WithOutput(w io.Writer) Option {
	return func(a *App) { a.Out = w }
}

// WithErrOutput sets the error output writer.
func WithErrOutput(w io.Writer) Option {
	return func(a *App) { a.ErrOut = w }
}

// WithFormat sets the default output format.
func WithFormat(f string) Option {
	return func(a *App) { a.Format = f }
}

// WithAPIURL sets the Sable API base URL.
func WithAPIURL(url string) Option {
	return func(a *App) { a.APIURL = url }
}

// WithOrgID sets the organization ID.
func WithOrgID(id string) Option {
	return func(a *App) { a.OrgID = id }
}

// WithToken sets the authentication token.
func WithToken(t string) Option {
	return func(a *App) { a.Token = t }
}

// WithVerbose enables verbose logging.
func WithVerbose(v bool) Option {
	return func(a *App) { a.Verbose = v }
}

// WithNoColor disables colored output.
func WithNoColor(v bool) Option {
	return func(a *App) { a.NoColor = v }
}
