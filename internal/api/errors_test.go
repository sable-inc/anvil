package api_test

import (
	"fmt"
	"testing"

	"github.com/sable-inc/anvil/internal/api"
)

func TestNewFromStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		message string
		hint    string
	}{
		{"404", 404, "not found", ""},
		{"401", 401, "unauthorized", "re-login"},
		{"403", 403, "forbidden", ""},
		{"409", 409, "conflict", ""},
		{"422", 422, "invalid", "check input"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := api.NewFromStatus(tt.status, tt.message, tt.hint)
			if err.Error() == "" {
				t.Error("Error() returned empty string")
			}
		})
	}
}

// TestAs_GenericTypeExtraction tests Go 1.26 errors.AsType via api.As[T].
func TestAs_GenericTypeExtraction(t *testing.T) {
	t.Run("extracts typed error", func(t *testing.T) {
		err := api.NewFromStatus(404, "not found", "try another id")
		nf, ok := api.As[*api.NotFoundError](err)
		if !ok {
			t.Fatal("As[*NotFoundError] should match a 404 error")
		}
		if nf.StatusCode != 404 {
			t.Errorf("StatusCode = %d, want 404", nf.StatusCode)
		}
		if nf.Hint != "try another id" {
			t.Errorf("Hint = %q, want %q", nf.Hint, "try another id")
		}
	})

	t.Run("does not match wrong type", func(t *testing.T) {
		err := api.NewFromStatus(404, "not found", "")
		_, ok := api.As[*api.UnauthorizedError](err)
		if ok {
			t.Error("As[*UnauthorizedError] should not match a 404 error")
		}
	})

	t.Run("works through wrapped errors", func(t *testing.T) {
		err := api.NewFromStatus(401, "unauthorized", "")
		wrapped := fmt.Errorf("operation failed: %w", err)

		ua, ok := api.As[*api.UnauthorizedError](wrapped)
		if !ok {
			t.Fatal("As[*UnauthorizedError] should work through fmt.Errorf wrapping")
		}
		if ua.StatusCode != 401 {
			t.Errorf("StatusCode = %d, want 401", ua.StatusCode)
		}
	})

	t.Run("all typed errors", func(t *testing.T) {
		cases := []struct {
			status int
			check  func(error) bool
			label  string
		}{
			{404, func(e error) bool { _, ok := api.As[*api.NotFoundError](e); return ok }, "NotFound"},
			{401, func(e error) bool { _, ok := api.As[*api.UnauthorizedError](e); return ok }, "Unauthorized"},
			{403, func(e error) bool { _, ok := api.As[*api.ForbiddenError](e); return ok }, "Forbidden"},
			{409, func(e error) bool { _, ok := api.As[*api.ConflictError](e); return ok }, "Conflict"},
			{422, func(e error) bool { _, ok := api.As[*api.ValidationError](e); return ok }, "Validation"},
		}
		for _, c := range cases {
			t.Run(c.label, func(t *testing.T) {
				err := api.NewFromStatus(c.status, "msg", "")
				if !c.check(err) {
					t.Errorf("As[%s] should match status %d", c.label, c.status)
				}
			})
		}
	})
}

func TestNewFromStatus_Unknown(t *testing.T) {
	err := api.NewFromStatus(500, "server error", "")

	_, notFound := api.As[*api.NotFoundError](err)
	_, unauth := api.As[*api.UnauthorizedError](err)
	if notFound || unauth {
		t.Error("500 should not match specific error types")
	}

	want := "API error 500: server error"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestResponseError_WithHint(t *testing.T) {
	err := api.NewFromStatus(401, "unauthorized", "re-login")
	want := "API error 401: unauthorized (hint: re-login)"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}
