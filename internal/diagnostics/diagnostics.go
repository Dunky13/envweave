// Package diagnostics provides opt-in, command-scoped progress on stderr.
// Callers must supply only public operational metadata, never credentials,
// command arguments, environments, request bodies, or unfiltered errors.
package diagnostics

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type contextKey struct{}

type output struct {
	level  int
	writer io.Writer
	mu     sync.Mutex
}

// With enables up to three cumulative levels: phases, details, and timings.
// A nil writer discards diagnostics. Child contexts retain cancellation.
func With(ctx context.Context, level int, out io.Writer) context.Context {
	if out == nil {
		out = io.Discard
	}
	return context.WithValue(ctx, contextKey{}, &output{level: min(3, max(0, level)), writer: out})
}

// Level returns zero when diagnostics have not been enabled.
func Level(ctx context.Context) int {
	if out, ok := ctx.Value(contextKey{}).(*output); ok {
		return out.level
	}
	return 0
}

// Printf writes one diagnostic line when its minimum level is enabled.
// Formats and arguments must be explicitly selected non-secret metadata.
func Printf(ctx context.Context, minimum int, format string, args ...any) {
	out, ok := ctx.Value(contextKey{}).(*output)
	if !ok || out.level == 0 || out.level < minimum {
		return
	}
	out.mu.Lock()
	defer out.mu.Unlock()
	fmt.Fprintf(out.writer, "hikyo: "+format+"\n", args...)
}

// Time returns a completion callback for a named operation at level three.
// Typical use: defer diagnostics.Time(ctx, "verify release")().
func Time(ctx context.Context, operation string) func() {
	if Level(ctx) < 3 {
		return func() {}
	}
	start := time.Now()
	return func() { Printf(ctx, 3, "%s elapsed=%s", operation, time.Since(start).Round(time.Millisecond)) }
}

// Transport reports request methods, destination origins and response status.
// URL paths, queries, userinfo, headers, bodies and error strings are omitted:
// even a failed transport error can contain an entire credential-bearing URL.
type Transport struct{ Base http.RoundTripper }

func (t Transport) base() http.RoundTripper {
	if t.Base == nil {
		return http.DefaultTransport
	}
	return t.Base
}

// CloseIdleConnections preserves http.Client's transport cleanup contract.
func (t Transport) CloseIdleConnections() {
	if closer, ok := t.base().(interface{ CloseIdleConnections() }); ok {
		closer.CloseIdleConnections()
	}
}

func (t Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	origin := req.URL.Scheme + "://" + req.URL.Host
	Printf(ctx, 2, "HTTP %s origin=%q", req.Method, origin)
	defer Time(ctx, "HTTP request to response headers")()
	response, err := t.base().RoundTrip(req)
	if err != nil {
		Printf(ctx, 2, "HTTP %s failed", req.Method)
	} else {
		Printf(ctx, 2, "HTTP %s status=%d", req.Method, response.StatusCode)
	}
	return response, err
}
