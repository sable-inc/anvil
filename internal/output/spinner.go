package output

import (
	"context"
	"fmt"
	"io"
	"time"
)

// PollConfig configures async job polling for --watch flags.
type PollConfig struct {
	// Interval between polls.
	Interval time.Duration
	// Timeout after which polling stops. Zero means no timeout.
	Timeout time.Duration
	// StatusFunc is called each tick. Returns the current status string,
	// whether the job is done, and any error.
	StatusFunc func(ctx context.Context) (status string, done bool, err error)
	// OnStatus is called after each poll with the current status.
	OnStatus func(status string)
}

// Poll repeatedly calls StatusFunc until done, timeout, or context cancellation.
func Poll(ctx context.Context, w io.Writer, cfg PollConfig) error {
	if cfg.Interval == 0 {
		cfg.Interval = 2 * time.Second
	}

	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		status, done, err := cfg.StatusFunc(ctx)
		if err != nil {
			return err
		}
		if cfg.OnStatus != nil {
			cfg.OnStatus(status)
		}
		if done {
			return nil
		}

		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("polling timed out (last status: %s)", status)
			}
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
