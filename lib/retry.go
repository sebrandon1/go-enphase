package lib

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/sebrandon1/go-toolbox/httpx"
	"github.com/sebrandon1/go-toolbox/retry"
)

// maxRetryDelay caps both the exponential backoff and any server-provided Retry-After.
const maxRetryDelay = 30 * time.Second

// maxRetryAttempts is the total number of attempts, including the first.
const maxRetryAttempts = 4

// retryBaseDelay is the first backoff interval.
const retryBaseDelay = 500 * time.Millisecond

// isRetryableStatus reports whether an HTTP status warrants a retry.
func isRetryableStatus(code int) bool {
	return code == http.StatusTooManyRequests || code >= 500
}

// noJitter disables backoff jitter, preserving the deterministic delays this
// client historically used.
func noJitter(d time.Duration) time.Duration { return d }

// retryOptions returns the retry policy for GET requests: up to four attempts
// on 429/5xx responses and transport errors, honoring Retry-After, with
// exponential backoff capped at maxRetryDelay. baseDelay is a parameter so
// tests can use a short delay.
func retryOptions(baseDelay time.Duration) []retry.HTTPOption {
	return []retry.HTTPOption{
		retry.WithMaxAttempts(maxRetryAttempts),
		retry.WithBaseDelay(baseDelay),
		retry.WithHTTPMaxDelay(maxRetryDelay),
		retry.WithRetryMethods(http.MethodGet),
		retry.WithRetryStatus(isRetryableStatus),
		retry.WithJitter(noJitter),
	}
}

// doGet executes a GET request with the retry policy applied. Non-retryable
// responses are returned for the caller to handle; exhausted retries are
// converted to the same user-facing errors as any other failed status.
func (c *Client) doGet(ctx context.Context, req *http.Request) (*http.Response, error) {
	resp, err := retry.Do(ctx, c.HTTPClient, req, retryOptions(retryBaseDelay)...)
	if err == nil {
		return resp, nil
	}

	var rateErr *httpx.RateLimitError
	if errors.As(err, &rateErr) {
		return nil, friendlyHTTPError(http.StatusTooManyRequests, nil)
	}
	var httpErr *httpx.HTTPError
	if errors.As(err, &httpErr) {
		return nil, friendlyHTTPError(httpErr.StatusCode, []byte(httpErr.Body))
	}
	var netErr *httpx.NetworkError
	if errors.As(err, &netErr) {
		return nil, netErr.Cause
	}
	return nil, err
}
