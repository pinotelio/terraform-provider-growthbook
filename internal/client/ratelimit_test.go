package client

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

// Issue 01: a 429 with no Retry-After is retried with backoff and eventually
// succeeds rather than surfacing as an error (which Terraform would read as the
// resource needing replacement).
func TestSend_RetriesOn429(t *testing.T) {
	var calls int32
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message":"Too many requests, limit to 60 per minute"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"project":{"id":"prj_1","name":"ok"}}`))
	})
	c.retry.MinBackoff = 0
	c.retry.MaxBackoff = 0

	if _, err := c.GetProject(context.Background(), "prj_1"); err != nil {
		t.Fatalf("unexpected error after 429 retries: %s", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}
}

// Issue 01: an explicit Retry-After must be honored in full and not shortened to
// MaxBackoff — otherwise the retry immediately re-hits the same rate limit.
func TestSend_HonorsRetryAfterBeyondMaxBackoff(t *testing.T) {
	var calls int32
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Header().Set("Retry-After", "1") // 1 second
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"project":{"id":"prj_1","name":"ok"}}`))
	})
	// MaxBackoff far below the Retry-After: with the capping bug the retry would
	// fire after ~1ms; honoring Retry-After it must wait ~1s.
	c.retry.MinBackoff = time.Millisecond
	c.retry.MaxBackoff = time.Millisecond

	start := time.Now()
	if _, err := c.GetProject(context.Background(), "prj_1"); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if elapsed := time.Since(start); elapsed < 700*time.Millisecond {
		t.Fatalf("expected the client to honor Retry-After (~1s), but it waited only %s", elapsed)
	}
}

func TestRateLimiter_SpacesRequests(t *testing.T) {
	const interval = 40 * time.Millisecond
	r := &rateLimiter{interval: interval}
	ctx := context.Background()

	start := time.Now()
	for i := 0; i < 3; i++ {
		r.wait(ctx) // first is immediate; next two each wait one interval
	}
	// Two spacing intervals between three requests.
	if elapsed := time.Since(start); elapsed < 2*interval-5*time.Millisecond {
		t.Fatalf("expected requests to be spaced by ~%s, total elapsed only %s", interval, elapsed)
	}
}

func TestRateLimiter_NilIsNoOp(t *testing.T) {
	var r *rateLimiter
	start := time.Now()
	r.wait(context.Background()) // must not panic, must not block
	if elapsed := time.Since(start); elapsed > 10*time.Millisecond {
		t.Fatalf("nil limiter should be a no-op, waited %s", elapsed)
	}
}

func TestNewRateLimiter_DisabledForNonPositive(t *testing.T) {
	if newRateLimiter(0) != nil {
		t.Error("expected nil limiter for 0 requests/min")
	}
	if newRateLimiter(-5) != nil {
		t.Error("expected nil limiter for negative requests/min")
	}
	if r := newRateLimiter(60); r == nil || r.interval != time.Second {
		t.Errorf("expected 60/min => 1s interval, got %+v", r)
	}
}

func TestRetryAfter_ParsesSecondsAndDate(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
	resp.Header.Set("Retry-After", "12")
	if got := retryAfter(resp); got != 12*time.Second {
		t.Errorf("expected 12s, got %s", got)
	}

	resp.Header.Set("Retry-After", time.Now().Add(2*time.Second).UTC().Format(http.TimeFormat))
	if got := retryAfter(resp); got <= 0 {
		t.Errorf("expected a positive duration from an HTTP-date Retry-After, got %s", got)
	}

	resp.Header.Del("Retry-After")
	if got := retryAfter(resp); got != 0 {
		t.Errorf("expected 0 when header absent, got %s", got)
	}
}
