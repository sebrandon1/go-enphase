package lib

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
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

// isTimeoutError reports whether a transport error is a timeout. The previous
// retry transport retried only timeout-class network errors, so keep that
// narrower policy rather than retrying every transport failure.
func isTimeoutError(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

// retryOptions returns the retry policy for GET requests: up to four attempts
// on 429/5xx responses and timeout-class transport errors, honoring
// Retry-After, with exponential backoff capped at maxRetryDelay. baseDelay is a
// parameter so tests can use a short delay.
func retryOptions(baseDelay time.Duration) []retry.HTTPOption {
	return []retry.HTTPOption{
		retry.WithMaxAttempts(maxRetryAttempts),
		retry.WithBaseDelay(baseDelay),
		retry.WithHTTPMaxDelay(maxRetryDelay),
		retry.WithRetryMethods(http.MethodGet),
		retry.WithRetryStatus(isRetryableStatus),
		retry.WithRetryTransport(isTimeoutError),
		retry.WithJitter(noJitter),
	}
}

// doGet executes a GET request with the retry policy applied. Non-retryable
// responses are returned for the caller to handle; exhausted retries are
// converted to the same user-facing errors as any other failed status.
func (c *Client) doGet(ctx context.Context, req *http.Request) (*http.Response, error) {
	// The previous retry transport lived inside a single client.Do, so
	// http.Client.Timeout bounded the entire retry sequence. retry.Do issues a
	// separate client.Do per attempt, so impose the same total budget here.
	if timeout := c.HTTPClient.Timeout; timeout > 0 {
		reqCtx, cancel := context.WithTimeout(ctx, timeout)
		resp, err := retry.Do(reqCtx, c.HTTPClient, req.WithContext(reqCtx), retryOptions(retryBaseDelay)...)
		if err != nil {
			cancel()
			return nil, c.retryError(err)
		}
		// The deadline must outlive this call: canceling it would abort the
		// response body. Release it when the caller closes or finishes it.
		resp.Body = &cancelOnCloseBody{ReadCloser: resp.Body, cancel: cancel}
		return resp, nil
	}

	resp, err := retry.Do(ctx, c.HTTPClient, req, retryOptions(retryBaseDelay)...)
	if err != nil {
		return nil, c.retryError(err)
	}
	return resp, nil
}

// retryError converts a retry-loop error into the error shape callers expect.
func (c *Client) retryError(err error) error {
	var rateErr *httpx.RateLimitError
	if errors.As(err, &rateErr) {
		return friendlyHTTPError(http.StatusTooManyRequests, nil)
	}
	var httpErr *httpx.HTTPError
	if errors.As(err, &httpErr) {
		return friendlyHTTPError(httpErr.StatusCode, []byte(httpErr.Body))
	}
	var netErr *httpx.NetworkError
	if errors.As(err, &netErr) {
		return netErr.Cause
	}
	return err
}

// cancelOnCloseBody releases the retry-sequence deadline once the caller is
// done with the response body.
type cancelOnCloseBody struct {
	io.ReadCloser
	once   sync.Once
	cancel context.CancelFunc
}

func (b *cancelOnCloseBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if err != nil {
		b.once.Do(b.cancel)
	}
	return n, err
}

func (b *cancelOnCloseBody) Close() error {
	err := b.ReadCloser.Close()
	b.once.Do(b.cancel)
	return err
}
