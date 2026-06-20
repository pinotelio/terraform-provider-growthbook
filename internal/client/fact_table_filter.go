package client

import (
	"context"
	"net/http"
	"net/url"
)

// FactTableFilter mirrors the GrowthBook FactTableFilter model. A filter is a
// reusable SQL expression scoped to a single fact table.
type FactTableFilter struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Value       string `json:"value"`
	ManagedBy   string `json:"managedBy,omitempty"`
	DateCreated string `json:"dateCreated,omitempty"`
	DateUpdated string `json:"dateUpdated,omitempty"`
}

// FactTableFilterInput is the create/update request body for a fact table
// filter. Both create and update share the same shape.
type FactTableFilterInput struct {
	Name        string  `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Value       string  `json:"value,omitempty"`
	ManagedBy   *string `json:"managedBy,omitempty"`
}

type factTableFilterEnvelope struct {
	FactTableFilter *FactTableFilter `json:"factTableFilter"`
}

func factTableFiltersPath(factTableID string) string {
	return "/fact-tables/" + url.PathEscape(factTableID) + "/filters"
}

// CreateFactTableFilter creates a filter on the given fact table.
func (c *Client) CreateFactTableFilter(ctx context.Context, factTableID string, in FactTableFilterInput) (*FactTableFilter, error) {
	var out factTableFilterEnvelope
	if err := c.doJSON(ctx, http.MethodPost, factTableFiltersPath(factTableID), in, &out); err != nil {
		return nil, err
	}
	return out.FactTableFilter, nil
}

// GetFactTableFilter fetches a single fact table filter by ID.
func (c *Client) GetFactTableFilter(ctx context.Context, factTableID, filterID string) (*FactTableFilter, error) {
	var out factTableFilterEnvelope
	if err := c.doJSON(ctx, http.MethodGet, factTableFiltersPath(factTableID)+"/"+url.PathEscape(filterID), nil, &out); err != nil {
		return nil, err
	}
	return out.FactTableFilter, nil
}

// UpdateFactTableFilter updates a fact table filter. The GrowthBook API uses
// POST for filter updates.
func (c *Client) UpdateFactTableFilter(ctx context.Context, factTableID, filterID string, in FactTableFilterInput) (*FactTableFilter, error) {
	var out factTableFilterEnvelope
	if err := c.doJSON(ctx, http.MethodPost, factTableFiltersPath(factTableID)+"/"+url.PathEscape(filterID), in, &out); err != nil {
		return nil, err
	}
	return out.FactTableFilter, nil
}

// DeleteFactTableFilter deletes a fact table filter.
func (c *Client) DeleteFactTableFilter(ctx context.Context, factTableID, filterID string) error {
	return c.doJSON(ctx, http.MethodDelete, factTableFiltersPath(factTableID)+"/"+url.PathEscape(filterID), nil, nil)
}
