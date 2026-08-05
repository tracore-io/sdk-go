package tracore

import (
	"context"
	"fmt"
	"time"
)

// PollOptions configures Runs.Poll and the Extract auto-poll. The zero value polls
// every second for up to 60 attempts.
type PollOptions struct {
	// Interval between poll requests. Default 1s.
	Interval time.Duration
	// MaxAttempts before giving up. Default 60.
	MaxAttempts int
	// OnProgress, if set, is called with the current run on each non-terminal poll.
	OnProgress func(*Run)
}

func isTerminal(s RunStatus) bool {
	return s == RunStatusCompleted || s == RunStatusFailed || s == RunStatusValidationFailed
}

// pollRun polls a run until it reaches a terminal status, ctx is cancelled, or the
// attempt budget is exhausted (returns a TracoreError with Status 408).
func pollRun(ctx context.Context, c *Client, runID string, opts *PollOptions) (*Run, error) {
	interval := time.Second
	maxAttempts := 60
	var onProgress func(*Run)
	if opts != nil {
		if opts.Interval > 0 {
			interval = opts.Interval
		}
		if opts.MaxAttempts > 0 {
			maxAttempts = opts.MaxAttempts
		}
		onProgress = opts.OnProgress
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		res, err := c.gen.GetRunByIdWithResponse(ctx, runID)
		if err != nil {
			return nil, transportError(err)
		}
		run, derr := decode(res.JSON200, res.StatusCode(), res.Body)
		if derr != nil {
			return nil, derr
		}
		if isTerminal(run.Status) {
			return run, nil
		}
		if onProgress != nil {
			onProgress(run)
		}
		select {
		case <-ctx.Done():
			return nil, transportError(ctx.Err())
		case <-time.After(interval):
		}
	}

	// Status 0 + a synthetic code: this is a client-side give-up, not an HTTP
	// response the server sent. The run itself is still live.
	return nil, &TracoreError{
		Code:    ErrCodePollingTimeout,
		Message: fmt.Sprintf("polling timed out after %d attempts for run %s", maxAttempts, runID),
	}
}
