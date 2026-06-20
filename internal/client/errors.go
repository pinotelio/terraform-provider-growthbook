package client

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned when the API responds with HTTP 404. Callers should
// use errors.Is(err, client.ErrNotFound) to detect a missing resource so they
// can remove it from Terraform state rather than surfacing a hard error.
var ErrNotFound = errors.New("growthbook: resource not found")

// APIError represents a non-success HTTP response from the GrowthBook API.
type APIError struct {
	StatusCode int
	Method     string
	Path       string
	// Message is the best-effort human readable error extracted from the
	// response body (GrowthBook returns {"message": "..."} on errors).
	Message string
	// Body is the raw response body, used as a fallback when Message is empty.
	Body string
}

func (e *APIError) Error() string {
	detail := e.Message
	if detail == "" {
		detail = e.Body
	}
	if detail == "" {
		return fmt.Sprintf("growthbook: %s %s returned HTTP %d", e.Method, e.Path, e.StatusCode)
	}
	return fmt.Sprintf("growthbook: %s %s returned HTTP %d: %s", e.Method, e.Path, e.StatusCode, detail)
}
