// Package client implements a typed HTTP client for the GrowthBook REST API
// (https://docs.growthbook.io/api). It is written for the Terraform provider
// and intentionally only models the surface the provider manages.
package client

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"
)

const defaultBaseURL = "https://api.growthbook.io/api/v1"

// RetryPolicy controls how transient API failures (HTTP 429 / 5xx and network
// errors) are retried with exponential backoff.
type RetryPolicy struct {
	MaxAttempts int
	MinBackoff  time.Duration
	MaxBackoff  time.Duration
	Multiplier  float64
}

func defaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 4,
		MinBackoff:  500 * time.Millisecond,
		MaxBackoff:  5 * time.Second,
		Multiplier:  2.0,
	}
}

// Client talks to a single GrowthBook instance. It is safe for concurrent use;
// mutating requests are serialized internally because several GrowthBook
// resources (attributes, environments, namespaces) are persisted as arrays and
// updated via read-modify-write on the server, which is not concurrency safe.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	retry      RetryPolicy
	pageLimit  int
	limiter    *rateLimiter

	writeMu sync.Mutex
}

// rateLimiter spaces outgoing requests to stay under a fixed per-minute budget.
// GrowthBook rate limits its REST API (self-hosted defaults to 60 requests per
// minute); without throttling, a refresh of many resources at Terraform's
// default parallelism bursts past the limit and trips HTTP 429. The limiter
// hands out evenly spaced "slots": concurrent callers queue and proceed one
// interval apart, smoothing bursts instead of rejecting them.
type rateLimiter struct {
	mu       sync.Mutex
	interval time.Duration
	next     time.Time
}

func newRateLimiter(perMinute int) *rateLimiter {
	if perMinute <= 0 {
		return nil
	}
	return &rateLimiter{interval: time.Minute / time.Duration(perMinute)}
}

// wait blocks until the limiter permits another request, or until ctx is
// cancelled. A nil limiter is a no-op, so callers need not branch on it.
func (r *rateLimiter) wait(ctx context.Context) {
	if r == nil {
		return
	}
	r.mu.Lock()
	now := time.Now()
	if r.next.Before(now) {
		r.next = now
	}
	wait := r.next.Sub(now)
	r.next = r.next.Add(r.interval)
	r.mu.Unlock()

	if wait <= 0 {
		return
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

// Option customizes a Client.
type Option func(*Client)

// New creates a GrowthBook API client. baseURL may be empty, in which case the
// public GrowthBook Cloud endpoint is used. A trailing slash on baseURL is
// trimmed so callers can pass paths beginning with "/".
func New(baseURL, apiKey string, opts ...Option) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	c := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: http.DefaultClient,
		retry:      defaultRetryPolicy(),
		pageLimit:  100,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithHTTPClient sets a custom *http.Client (for timeouts, TLS settings, etc).
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.httpClient = h
		}
	}
}

// WithRetryPolicy overrides the default backoff configuration.
func WithRetryPolicy(p RetryPolicy) Option {
	return func(c *Client) {
		if p.MaxAttempts < 1 {
			p.MaxAttempts = 1
		}
		if p.Multiplier < 1 {
			p.Multiplier = 2.0
		}
		c.retry = p
	}
}

// WithRateLimit caps how many requests the client issues per minute by spacing
// them evenly. A value <= 0 disables throttling (the default). Set it at or
// below the GrowthBook instance's documented limit (self-hosted: 60/min) to keep
// large refreshes from tripping HTTP 429.
func WithRateLimit(perMinute int) Option {
	return func(c *Client) {
		c.limiter = newRateLimiter(perMinute)
	}
}

// WithPageLimit sets how many items are requested per page for list endpoints.
func WithPageLimit(limit int) Option {
	return func(c *Client) {
		if limit > 0 {
			c.pageLimit = limit
		}
	}
}

// PageLimit returns the configured page size for paginated list requests.
func (c *Client) PageLimit() int { return c.pageLimit }

func redactKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:4] + "…" + apiKey[len(apiKey)-4:]
}
