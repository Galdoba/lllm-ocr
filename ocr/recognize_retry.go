package ocr

// RetryState holds per-page retry state.
type RetryState struct {
	CurrentRetry   int
	MaxRetries     int
	OriginalArgs   []string
	RetryTempArgs  []string
}

// CanRetry returns true if we can attempt another retry.
// MaxRetries=N means N retries after the first attempt (N+1 total attempts).
func (s *RetryState) CanRetry() bool {
	return s.CurrentRetry <= s.MaxRetries
}

// IncrementRetry increments the current retry counter.
func (s *RetryState) IncrementRetry() {
	s.CurrentRetry++
}

// NewRetryState creates a new RetryState for a page.
func NewRetryState(maxRetries int, originalArgs, retryTempArgs []string) *RetryState {
	return &RetryState{
		CurrentRetry:   0,
		MaxRetries:     maxRetries,
		OriginalArgs:   originalArgs,
		RetryTempArgs:  retryTempArgs,
	}
}
