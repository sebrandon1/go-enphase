package lib

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"
)

// maxRetryDelay caps both the exponential backoff and any server-provided Retry-After.
const maxRetryDelay = 30 * time.Second

// retryTransport is an http.RoundTripper that retries GET requests on 429 and 5xx
// responses (honoring Retry-After) and transient network errors with exponential backoff.
type retryTransport struct {
	inner     http.RoundTripper
	maxTries  int
	baseDelay time.Duration
}

// RoundTrip implements http.RoundTripper with retry logic for GET requests.
func (rt *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method != http.MethodGet {
		return rt.inner.RoundTrip(req)
	}

	var (
		resp  *http.Response
		err   error
		delay = rt.baseDelay
		wait  = rt.baseDelay
	)

	for attempt := 0; attempt <= rt.maxTries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(wait):
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
			delay = min(delay*2, maxRetryDelay)
			wait = delay
		}

		resp, err = rt.inner.RoundTrip(req)
		if attempt == rt.maxTries || !shouldRetry(resp, err) {
			return resp, err
		}
		if resp != nil {
			if d, ok := retryAfter(resp); ok {
				wait = d
			}
			drainAndClose(resp.Body)
		}
	}

	return resp, err
}

// shouldRetry reports whether a response or error warrants a retry.
func shouldRetry(resp *http.Response, err error) bool {
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false
		}
		var netErr net.Error
		return errors.As(err, &netErr) && netErr.Timeout()
	}
	return resp != nil && (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500)
}

// retryAfter parses a Retry-After header given in seconds, capped at maxRetryDelay.
// HTTP-date values and invalid or negative values are ignored.
func retryAfter(resp *http.Response) (time.Duration, bool) {
	if resp == nil {
		return 0, false
	}
	secs, err := strconv.Atoi(resp.Header.Get("Retry-After"))
	if err != nil || secs < 0 {
		return 0, false
	}
	return min(time.Duration(secs)*time.Second, maxRetryDelay), true
}
