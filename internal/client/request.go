package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// doJSON performs an authenticated JSON request against the GrowthBook API and,
// on success, unmarshals the response body into out (which may be nil to
// discard the body). A 404 is returned as ErrNotFound; any other non-2xx status
// is returned as *APIError. Mutating verbs are serialized via writeMu.
func (c *Client) doJSON(ctx context.Context, method, path string, body, out any) error {
	if method != http.MethodGet {
		c.writeMu.Lock()
		defer c.writeMu.Unlock()
	}

	var payload []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}
		payload = b
	}

	fullURL := c.baseURL + path
	resp, respBody, err := c.send(ctx, method, fullURL, payload)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Method:     method,
			Path:       path,
			Message:    extractMessage(respBody),
			Body:       string(bytes.TrimSpace(respBody)),
		}
	}

	if out == nil || len(bytes.TrimSpace(respBody)) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decoding response from %s %s: %w", method, path, err)
	}
	return nil
}

// send issues the HTTP request with retry/backoff on 429 and 5xx responses and
// transport errors. It returns the final response (with body already drained)
// and the raw body bytes.
func (c *Client) send(ctx context.Context, method, fullURL string, payload []byte) (*http.Response, []byte, error) {
	interval := c.retry.MinBackoff
	var lastErr error

	for attempt := 1; attempt <= c.retry.MaxAttempts; attempt++ {
		// Honor the client-side rate limit (if configured) before every attempt,
		// retries included, so backoff and throttling compose.
		c.limiter.wait(ctx)
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}

		var reader io.Reader
		if payload != nil {
			reader = bytes.NewReader(payload)
		}
		req, err := http.NewRequestWithContext(ctx, method, fullURL, reader)
		if err != nil {
			return nil, nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		tflog.Debug(ctx, "growthbook API request", map[string]any{
			"method":  method,
			"url":     fullURL,
			"attempt": attempt,
			"apiKey":  redactKey(c.apiKey),
		})

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt == c.retry.MaxAttempts {
				return nil, nil, err
			}
			c.sleep(ctx, minDuration(interval, c.retry.MaxBackoff), attempt, 0, err)
			interval = minDuration(time.Duration(float64(interval)*c.retry.Multiplier), c.retry.MaxBackoff)
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		tflog.Debug(ctx, "growthbook API response", map[string]any{
			"method": method,
			"url":    fullURL,
			"status": resp.StatusCode,
		})

		retryable := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
		if !retryable || attempt == c.retry.MaxAttempts {
			return resp, respBody, nil
		}

		// Exponential backoff is bounded by MaxBackoff, but an explicit
		// Retry-After from the server is honored in full: a 429 that asks us to
		// wait (e.g.) 30s must not be shortened to MaxBackoff, or the retry just
		// hits the same limit again.
		wait := minDuration(interval, c.retry.MaxBackoff)
		if resp.StatusCode == http.StatusTooManyRequests {
			if ra := retryAfter(resp); ra > 0 {
				wait = ra
			}
		}
		c.sleep(ctx, wait, attempt, resp.StatusCode, nil)
		interval = minDuration(time.Duration(float64(interval)*c.retry.Multiplier), c.retry.MaxBackoff)
	}

	if lastErr != nil {
		return nil, nil, lastErr
	}
	return nil, nil, fmt.Errorf("growthbook: request to %s exhausted retries", fullURL)
}

// sleep waits for the given duration (or until ctx is cancelled), logging the
// pending retry. The caller is responsible for advancing the exponential
// backoff interval between attempts.
func (c *Client) sleep(ctx context.Context, wait time.Duration, attempt, status int, cause error) {
	tflog.Warn(ctx, "growthbook API retry", map[string]any{
		"attempt": attempt,
		"status":  status,
		"wait_ms": wait.Milliseconds(),
		"error":   errString(cause),
	})

	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func retryAfter(resp *http.Response) time.Duration {
	v := resp.Header.Get("Retry-After")
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// extractMessage pulls a human readable error string out of a GrowthBook error
// body, which is typically {"message":"..."}.
func extractMessage(body []byte) string {
	var env struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return ""
	}
	if env.Message != "" {
		return env.Message
	}
	return env.Error
}

// unmarshal is a thin wrapper over json.Unmarshal used by list decoders.
func unmarshal(b []byte, v any) error {
	return json.Unmarshal(b, v)
}

// pagination is the common envelope GrowthBook returns alongside list payloads.
type pagination struct {
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
	Count      int  `json:"count"`
	Total      int  `json:"total"`
	HasMore    bool `json:"hasMore"`
	NextOffset *int `json:"nextOffset"`
}

// fetchAll repeatedly calls a paginated list endpoint at basePath and collects
// every page into a single slice. extract is given each decoded page's items.
func fetchAll[T any](ctx context.Context, c *Client, basePath string, decodePage func([]byte) ([]T, pagination, error)) ([]T, error) {
	var all []T
	offset := 0
	for {
		var raw json.RawMessage
		sep := "?"
		if containsQuery(basePath) {
			sep = "&"
		}
		path := basePath + sep + "limit=" + strconv.Itoa(c.pageLimit) + "&offset=" + strconv.Itoa(offset)
		if err := c.doJSON(ctx, http.MethodGet, path, nil, &raw); err != nil {
			return nil, err
		}
		items, page, err := decodePage(raw)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
		if !page.HasMore || page.NextOffset == nil {
			break
		}
		offset = *page.NextOffset
	}
	return all, nil
}

func containsQuery(p string) bool {
	for i := 0; i < len(p); i++ {
		if p[i] == '?' {
			return true
		}
	}
	return false
}
