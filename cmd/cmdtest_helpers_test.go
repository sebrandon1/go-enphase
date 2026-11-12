package cmd

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/sebrandon1/go-enphase/lib"
	"github.com/spf13/cobra"
)

const testSystemID = "123"

// routes maps a request path to a canned JSON body. Unknown paths return 404.
type routes map[string]string

func (r routes) handler(status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, ok := r[req.URL.Path]
		if !ok {
			http.NotFound(w, req)
			return
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})
}

// useCloud points the cloud client global at a mock server for the test.
// A non-200 status makes every known route fail with that status.
func useCloud(t *testing.T, r routes, status int) {
	t.Helper()
	srv := httptest.NewServer(r.handler(status))
	origURL, origClient, origSys, origFmt := lib.CloudBaseURL, cloudClient, systemID, outputFormat
	lib.CloudBaseURL = srv.URL
	client, err := lib.NewClient("key", "token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	cloudClient = client
	systemID = testSystemID
	outputFormat = "text"
	t.Cleanup(func() {
		srv.Close()
		lib.CloudBaseURL, cloudClient, systemID, outputFormat = origURL, origClient, origSys, origFmt
	})
}

// useEnvoy points the local Envoy client global at a mock TLS server.
func useEnvoy(t *testing.T, r routes, status int) {
	t.Helper()
	srv := httptest.NewTLSServer(r.handler(status))
	origClient, origFmt := localClient, outputFormat
	localClient = &lib.Client{
		EnvoyIP:    strings.TrimPrefix(srv.URL, "https://"),
		HTTPClient: srv.Client(),
	}
	outputFormat = "text"
	t.Cleanup(func() {
		srv.Close()
		localClient, outputFormat = origClient, origFmt
	})
}

// runCmd runs c.RunE and returns what it wrote to stdout plus its error.
func runCmd(t *testing.T, c *cobra.Command) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	runErr := c.RunE(c, nil)
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String(), runErr
}
