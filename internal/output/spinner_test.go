package output_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sable-inc/anvil/internal/output"
)

func TestPoll_ImmediateDone(t *testing.T) {
	var buf bytes.Buffer
	err := output.Poll(context.Background(), &buf, output.PollConfig{
		Interval: 10 * time.Millisecond,
		StatusFunc: func(_ context.Context) (string, bool, error) {
			return "done", true, nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPoll_MultiplePolls(t *testing.T) {
	var buf bytes.Buffer
	calls := 0
	err := output.Poll(context.Background(), &buf, output.PollConfig{
		Interval: 10 * time.Millisecond,
		StatusFunc: func(_ context.Context) (string, bool, error) {
			calls++
			if calls >= 3 {
				return "complete", true, nil
			}
			return "working", false, nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls < 3 {
		t.Errorf("expected at least 3 calls, got %d", calls)
	}
}

func TestPoll_OnStatus(t *testing.T) {
	var buf bytes.Buffer
	var statuses []string
	calls := 0
	err := output.Poll(context.Background(), &buf, output.PollConfig{
		Interval: 10 * time.Millisecond,
		StatusFunc: func(_ context.Context) (string, bool, error) {
			calls++
			if calls == 1 {
				return "pending", false, nil
			}
			if calls == 2 {
				return "building", false, nil
			}
			return "succeeded", true, nil
		},
		OnStatus: func(status string) {
			statuses = append(statuses, status)
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 3 {
		t.Errorf("expected 3 status callbacks, got %d: %v", len(statuses), statuses)
	}
}

func TestPoll_Error(t *testing.T) {
	var buf bytes.Buffer
	expectedErr := errors.New("api error")
	err := output.Poll(context.Background(), &buf, output.PollConfig{
		Interval: 10 * time.Millisecond,
		StatusFunc: func(_ context.Context) (string, bool, error) {
			return "", false, expectedErr
		},
	})
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestPoll_Timeout(t *testing.T) {
	var buf bytes.Buffer
	err := output.Poll(context.Background(), &buf, output.PollConfig{
		Interval: 10 * time.Millisecond,
		Timeout:  50 * time.Millisecond,
		StatusFunc: func(_ context.Context) (string, bool, error) {
			return "still working", false, nil
		},
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		// Our custom error wraps it
		if !contains(err.Error(), "polling timed out") {
			t.Errorf("expected timeout error, got: %v", err)
		}
	}
}

func TestPoll_ContextCanceled(t *testing.T) {
	var buf bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := output.Poll(ctx, &buf, output.PollConfig{
		Interval: 10 * time.Millisecond,
		StatusFunc: func(_ context.Context) (string, bool, error) {
			calls++
			if calls >= 2 {
				cancel()
			}
			return "working", false, nil
		},
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
