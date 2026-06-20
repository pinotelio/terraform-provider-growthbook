package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New(srv.URL, "test-key")
}

func TestDoJSON_DecodesEnvelopeAndSendsAuth(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("missing/incorrect auth header: %q", got)
		}
		if r.URL.Path != "/projects/prj_1" {
			t.Errorf("unexpected path: %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"project":{"id":"prj_1","name":"Demo"}}`))
	})

	project, err := c.GetProject(context.Background(), "prj_1")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if project.ID != "prj_1" || project.Name != "Demo" {
		t.Fatalf("unexpected project: %+v", project)
	}
}

func TestDoJSON_NotFoundMapsToErrNotFound(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})

	_, err := c.GetProject(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDoJSON_APIErrorIncludesMessage(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"name is required"}`))
	})

	_, err := c.CreateProject(context.Background(), ProjectInput{})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "name is required" {
		t.Errorf("expected message to be extracted, got %q", apiErr.Message)
	}
}

func TestSend_RetriesOnServerError(t *testing.T) {
	var calls int
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"project":{"id":"prj_1","name":"ok"}}`))
	})
	c.retry.MinBackoff = 0
	c.retry.MaxBackoff = 0

	if _, err := c.GetProject(context.Background(), "prj_1"); err != nil {
		t.Fatalf("unexpected error after retries: %s", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 attempts, got %d", calls)
	}
}
