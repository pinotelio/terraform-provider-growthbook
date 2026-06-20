package client

import (
	"context"
	"net/http"
	"net/url"
)

// ProjectSettings holds the optional statistics configuration of a project.
type ProjectSettings struct {
	StatsEngine     string   `json:"statsEngine,omitempty"`
	ConfidenceLevel *float64 `json:"confidenceLevel,omitempty"`
	PValueThreshold *float64 `json:"pValueThreshold,omitempty"`
}

// Project mirrors the GrowthBook Project model.
type Project struct {
	ID          string          `json:"id,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Settings    ProjectSettings `json:"settings,omitempty"`
	DateCreated string          `json:"dateCreated,omitempty"`
	DateUpdated string          `json:"dateUpdated,omitempty"`
}

// ProjectInput is the create/update request body for a project.
type ProjectInput struct {
	Name        string           `json:"name,omitempty"`
	Description *string          `json:"description,omitempty"`
	Settings    *ProjectSettings `json:"settings,omitempty"`
}

type projectEnvelope struct {
	Project *Project `json:"project"`
}

// CreateProject creates a project.
func (c *Client) CreateProject(ctx context.Context, in ProjectInput) (*Project, error) {
	var out projectEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/projects", in, &out); err != nil {
		return nil, err
	}
	return out.Project, nil
}

// GetProject fetches a project by ID.
func (c *Client) GetProject(ctx context.Context, id string) (*Project, error) {
	var out projectEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/projects/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.Project, nil
}

// UpdateProject updates a project by ID.
func (c *Client) UpdateProject(ctx context.Context, id string, in ProjectInput) (*Project, error) {
	var out projectEnvelope
	if err := c.doJSON(ctx, http.MethodPut, "/projects/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return out.Project, nil
}

// DeleteProject deletes a project by ID.
func (c *Client) DeleteProject(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/projects/"+url.PathEscape(id), nil, nil)
}

// ListProjects returns every project, following pagination.
func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	return fetchAll(ctx, c, "/projects", func(b []byte) ([]Project, pagination, error) {
		var page struct {
			Projects []Project `json:"projects"`
			pagination
		}
		if err := unmarshal(b, &page); err != nil {
			return nil, pagination{}, err
		}
		return page.Projects, page.pagination, nil
	})
}
