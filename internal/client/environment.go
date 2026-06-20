package client

import (
	"context"
	"net/http"
	"net/url"
)

// Environment mirrors the GrowthBook Environment model. The ID is user supplied
// at creation time and acts as the environment's name.
type Environment struct {
	ID           string   `json:"id"`
	Description  string   `json:"description"`
	ToggleOnList bool     `json:"toggleOnList"`
	DefaultState bool     `json:"defaultState"`
	Projects     []string `json:"projects"`
	Parent       string   `json:"parent,omitempty"`
}

// EnvironmentCreateInput is the POST /environments body.
type EnvironmentCreateInput struct {
	ID           string   `json:"id"`
	Description  *string  `json:"description,omitempty"`
	ToggleOnList *bool    `json:"toggleOnList,omitempty"`
	DefaultState *bool    `json:"defaultState,omitempty"`
	Projects     []string `json:"projects,omitempty"`
	Parent       *string  `json:"parent,omitempty"`
}

// EnvironmentUpdateInput is the PUT /environments/{id} body (id is immutable).
type EnvironmentUpdateInput struct {
	Description  *string  `json:"description,omitempty"`
	ToggleOnList *bool    `json:"toggleOnList,omitempty"`
	DefaultState *bool    `json:"defaultState,omitempty"`
	Projects     []string `json:"projects,omitempty"`
}

type environmentEnvelope struct {
	Environment *Environment `json:"environment"`
}

// CreateEnvironment creates an environment.
func (c *Client) CreateEnvironment(ctx context.Context, in EnvironmentCreateInput) (*Environment, error) {
	var out environmentEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/environments", in, &out); err != nil {
		return nil, err
	}
	return out.Environment, nil
}

// UpdateEnvironment updates an environment by ID.
func (c *Client) UpdateEnvironment(ctx context.Context, id string, in EnvironmentUpdateInput) (*Environment, error) {
	var out environmentEnvelope
	if err := c.doJSON(ctx, http.MethodPut, "/environments/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return out.Environment, nil
}

// DeleteEnvironment deletes an environment by ID.
func (c *Client) DeleteEnvironment(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/environments/"+url.PathEscape(id), nil, nil)
}

// ListEnvironments returns all environments. The endpoint is not paginated.
func (c *Client) ListEnvironments(ctx context.Context) ([]Environment, error) {
	var out struct {
		Environments []Environment `json:"environments"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/environments", nil, &out); err != nil {
		return nil, err
	}
	return out.Environments, nil
}

// GetEnvironment returns a single environment by ID. GrowthBook has no
// get-by-id endpoint for environments, so this filters the list.
func (c *Client) GetEnvironment(ctx context.Context, id string) (*Environment, error) {
	envs, err := c.ListEnvironments(ctx)
	if err != nil {
		return nil, err
	}
	for i := range envs {
		if envs[i].ID == id {
			return &envs[i], nil
		}
	}
	return nil, ErrNotFound
}
