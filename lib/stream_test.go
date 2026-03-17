package lib

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// sseTestServer creates a TLS test server that sends the provided SSE data lines and
// then optionally blocks until the provided channel is closed.
func sseTestServer(t *testing.T, events []string, block <-chan struct{}) *httptest.Server {
	t.Helper()
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Error("ResponseWriter does not support Flusher")
			return
		}
		for _, ev := range events {
			fmt.Fprintf(w, "data: %s\n\n", ev)
			flusher.Flush()
		}
		if block != nil {
			<-block
		}
	}))
}

// envoyClientForTLSServer creates a Client wired to the given TLS test server.
func envoyClientForTLSServer(server *httptest.Server) *Client {
	hostPort := server.Listener.Addr().String()
	return &Client{
		EnvoyIP: hostPort,
		HTTPClient: &http.Client{
			Transport: &retryTransport{
				inner:     server.Client().Transport,
				maxTries:  3,
				baseDelay: 500 * time.Millisecond,
			},
		},
	}
}

func TestStreamMeterMissingEnvoyIP(t *testing.T) {
	client := &Client{HTTPClient: newHTTPClientWithTLS(false)}
	var gotErr error
	client.StreamMeter(context.Background(), func(ev *StreamMeterEvent, err error) {
		if gotErr == nil {
			gotErr = err
		}
	})
	if gotErr == nil {
		t.Error("Expected error for missing envoy IP, got nil")
	}
}

func TestStreamMeterSuccess(t *testing.T) {
	events := []string{
		`{"eid":1,"timestamp":1000,"activePower":100.0}`,
		`{"eid":1,"timestamp":1001,"activePower":200.0}`,
		`{"eid":1,"timestamp":1002,"activePower":300.0}`,
	}
	server := sseTestServer(t, events, nil)
	defer server.Close()

	client := envoyClientForTLSServer(server)

	var received []*StreamMeterEvent
	err := client.streamMeterOnce(context.Background(), func(ev *StreamMeterEvent, e error) {
		if e == nil {
			received = append(received, ev)
		}
	})
	if err != nil {
		t.Fatalf("streamMeterOnce returned error: %v", err)
	}
	if len(received) != 3 {
		t.Fatalf("Expected 3 events, got %d", len(received))
	}
	if received[0].ActPower != 100.0 {
		t.Errorf("Expected first event ActPower 100.0, got %f", received[0].ActPower)
	}
	if received[2].ActPower != 300.0 {
		t.Errorf("Expected third event ActPower 300.0, got %f", received[2].ActPower)
	}
}

func TestStreamMeterContextCancel(t *testing.T) {
	// Server sends events in a loop until the client disconnects.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		for {
			select {
			case <-r.Context().Done():
				return
			default:
			}
			fmt.Fprintf(w, "data: {\"eid\":1,\"timestamp\":1000,\"activePower\":100.0}\n\n")
			flusher.Flush()
			select {
			case <-r.Context().Done():
				return
			case <-time.After(20 * time.Millisecond):
			}
		}
	}))
	defer server.Close()

	client := envoyClientForTLSServer(server)
	ctx, cancel := context.WithCancel(context.Background())

	received := 0
	done := make(chan struct{})
	go func() {
		defer close(done)
		client.StreamMeter(ctx, func(ev *StreamMeterEvent, err error) {
			if ev != nil {
				received++
				cancel()
			}
		})
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("StreamMeter did not return after context cancel")
	}
	if received < 1 {
		t.Errorf("Expected at least 1 event before cancel, got %d", received)
	}
}

func TestStreamMeterMalformedEvent(t *testing.T) {
	events := []string{
		`not valid json`,
		`{"eid":1,"timestamp":1000,"activePower":100.0}`,
	}
	server := sseTestServer(t, events, nil)
	defer server.Close()

	client := envoyClientForTLSServer(server)

	var mu sync.Mutex
	var goodCount, errCount int
	err := client.streamMeterOnce(context.Background(), func(ev *StreamMeterEvent, e error) {
		mu.Lock()
		defer mu.Unlock()
		if e != nil {
			errCount++
		} else {
			goodCount++
		}
	})
	if err != nil {
		t.Fatalf("streamMeterOnce returned unexpected error: %v", err)
	}
	if goodCount != 1 {
		t.Errorf("Expected 1 good event, got %d", goodCount)
	}
	if errCount != 1 {
		t.Errorf("Expected 1 parse error, got %d", errCount)
	}
}

func TestStreamMeterReconnect(t *testing.T) {
	var mu sync.Mutex
	attempts := 0
	events := []string{`{"eid":1,"timestamp":1000,"activePower":100.0}`}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		attempts++
		attempt := attempts
		mu.Unlock()

		if attempt == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		for _, ev := range events {
			fmt.Fprintf(w, "data: %s\n\n", ev)
			flusher.Flush()
		}
	}))
	defer server.Close()

	client := envoyClientForTLSServer(server)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	received := 0
	done := make(chan struct{})
	go func() {
		defer close(done)
		client.StreamMeter(ctx, func(ev *StreamMeterEvent, err error) {
			if ev != nil {
				received++
				cancel()
			}
		})
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Error("StreamMeter did not receive event after reconnect")
	}
	if received != 1 {
		t.Errorf("Expected 1 event after reconnect, got %d", received)
	}
	if attempts < 2 {
		t.Errorf("Expected at least 2 attempts (reconnect), got %d", attempts)
	}
}
