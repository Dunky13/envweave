package diagnostics

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

func TestDiagnosticLevelsAndCancellation(t *testing.T) {
	for level := 0; level <= 3; level++ {
		var out bytes.Buffer
		ctx := With(t.Context(), level, &out)
		for minimum := 1; minimum <= 3; minimum++ {
			Printf(ctx, minimum, "level%d", minimum)
		}
		if count := strings.Count(out.String(), "hikyo:"); count != level {
			t.Fatalf("level=%d output=%q", level, out.String())
		}
		if ctx.Done() != t.Context().Done() {
			t.Fatal("context cancellation lost")
		}
	}
	Printf(t.Context(), 1, "disabled")
	Printf(With(t.Context(), 3, nil), 1, "discarded")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestTransportDiagnosticsExcludeCredentialsAndPayloads(t *testing.T) {
	for _, fail := range []bool{false, true} {
		var out bytes.Buffer
		ctx := With(t.Context(), 3, &out)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://username:password@example.com/secret-path?token=secret-query#secret-fragment", strings.NewReader("secret-body"))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer secret-header")
		transport := Transport{Base: roundTripFunc(func(received *http.Request) (*http.Response, error) {
			if received != req {
				t.Fatal("diagnostic transport changed request")
			}
			if fail {
				return nil, errors.New("secret-transport-error")
			}
			return &http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader("secret-response"))}, nil
		})}
		response, err := transport.RoundTrip(req)
		if fail != (err != nil) {
			t.Fatalf("error=%v", err)
		}
		if response != nil {
			response.Body.Close()
		}
		for _, forbidden := range []string{"username", "password", "secret-", "Authorization", "token="} {
			if strings.Contains(out.String(), forbidden) {
				t.Fatalf("credential leaked: %s", out.String())
			}
		}
		if !strings.Contains(out.String(), "https://example.com") || !strings.Contains(out.String(), "elapsed=") {
			t.Fatalf("missing safe detail: %s", out.String())
		}
		if !fail && !strings.Contains(out.String(), "status=201") {
			t.Fatalf("missing status: %s", out.String())
		}
	}
}

func TestConcurrentDiagnosticsKeepCompleteLines(t *testing.T) {
	var out bytes.Buffer
	ctx := With(t.Context(), 1, &out)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Go(func() { Printf(ctx, 1, "phase") })
	}
	wg.Wait()
	if out.String() != strings.Repeat("hikyo: phase\n", 20) {
		t.Fatalf("interleaved diagnostics: %q", out.String())
	}
}
