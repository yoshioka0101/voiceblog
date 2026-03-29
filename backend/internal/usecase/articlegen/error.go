package articlegen

import (
	"errors"
	"fmt"
	"time"
)

// RateLimitError represents a temporary upstream quota/rate-limit failure during article generation.
type RateLimitError struct {
	RetryAfter time.Duration
	cause      error
}

func NewRateLimitError(retryAfter time.Duration, cause error) *RateLimitError {
	return &RateLimitError{
		RetryAfter: retryAfter,
		cause:      cause,
	}
}

func (e *RateLimitError) Error() string {
	if e == nil {
		return "article generation is temporarily rate limited; retry later"
	}
	if rounded := e.RetryAfter.Round(time.Second); rounded > 0 {
		return fmt.Sprintf("article generation is temporarily rate limited; retry in %s", rounded)
	}
	return "article generation is temporarily rate limited; retry later"
}

func (e *RateLimitError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func AsRateLimitError(err error) (*RateLimitError, bool) {
	var target *RateLimitError
	if !errors.As(err, &target) {
		return nil, false
	}
	return target, true
}
