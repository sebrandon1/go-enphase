package lib

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sebrandon1/go-toolbox/httpx"
)

// ErrStreamUnauthorized is returned and passed to the handler when the Envoy
// rejects the stream request with 401 or 403. StreamMeter stops retrying on
// this error.
var ErrStreamUnauthorized = errors.New("stream: unauthorized (401/403)")

// sseMaxLineBytes is the maximum SSE line the scanner will accept (1 MiB).
const sseMaxLineBytes = 1 << 20

// StreamMeter connects to the Envoy real-time meter SSE stream at /stream/meter and
// calls handler for each event. It automatically reconnects with exponential backoff
// (1s to 60s) on error; the backoff resets to 1s after a connection that delivers at
// least one event. On ErrStreamUnauthorized (HTTP 401/403) the handler is called with
// the error and the loop stops without reconnecting. The loop runs until ctx is canceled.
func (c *Client) StreamMeter(ctx context.Context, handler func(*StreamMeterEvent, error)) {
	if c.EnvoyIP == "" {
		handler(nil, fmt.Errorf("envoy IP is required"))
		return
	}

	delay := time.Second
	for {
		if ctx.Err() != nil {
			return
		}

		gotEvent := false
		wrapped := func(ev *StreamMeterEvent, err error) {
			if ev != nil {
				gotEvent = true
			}
			handler(ev, err)
		}

		err := c.streamMeterOnce(ctx, wrapped)
		if ctx.Err() != nil {
			return
		}

		if errors.Is(err, ErrStreamUnauthorized) {
			handler(nil, err)
			return
		}

		if gotEvent {
			delay = time.Second
		}

		if err != nil {
			handler(nil, fmt.Errorf("stream disconnected, reconnecting in %s: %w", delay, err))
		}
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return
		}
		delay = min(delay*2, 60*time.Second)
	}
}

// streamMeterOnce runs a single SSE connection to /stream/meter, calling handler for
// each parsed event. It returns when the connection closes or ctx is canceled.
func (c *Client) streamMeterOnce(ctx context.Context, handler func(*StreamMeterEvent, error)) error {
	url := "https://" + c.EnvoyIP + "/stream/meter"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")
	c.setEnvoyHeaders(req)

	streamClient := c.newStreamClient()
	resp, err := streamClient.Do(req)
	if err != nil {
		return err
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return ErrStreamUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		body := httpx.ReadErrorBody(resp.Body, 512)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 4096), sseMaxLineBytes)
	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil
		}
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var event StreamMeterEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			handler(nil, fmt.Errorf("failed to parse SSE event: %w", err))
			continue
		}
		handler(&event, nil)
	}

	return scanner.Err()
}

// newStreamClient returns an HTTP client suitable for SSE streaming: no timeout,
// but reuses the same TLS transport (without the retry wrapper).
func (c *Client) newStreamClient() *http.Client {
	inner := innerTransport(c.HTTPClient.Transport)
	var transport http.RoundTripper
	if inner != nil {
		transport = inner
	} else {
		transport = http.DefaultTransport
	}
	return &http.Client{Transport: transport}
}
