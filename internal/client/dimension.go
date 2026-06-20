package client

import (
	"context"
	"net/http"
	"net/url"
)

// Dimension mirrors the GrowthBook Dimension model.
type Dimension struct {
	ID             string `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	Owner          string `json:"owner,omitempty"`
	Description    string `json:"description,omitempty"`
	DatasourceID   string `json:"datasourceId,omitempty"`
	IdentifierType string `json:"identifierType,omitempty"`
	Query          string `json:"query,omitempty"`
	ManagedBy      string `json:"managedBy,omitempty"`
	DateCreated    string `json:"dateCreated,omitempty"`
	DateUpdated    string `json:"dateUpdated,omitempty"`
}

// DimensionInput is the create/update request body for a dimension. Pointer
// fields are omitted from the JSON when nil so server defaults apply.
type DimensionInput struct {
	Name           string  `json:"name,omitempty"`
	Description    *string `json:"description,omitempty"`
	Owner          *string `json:"owner,omitempty"`
	DatasourceID   string  `json:"datasourceId,omitempty"`
	IdentifierType string  `json:"identifierType,omitempty"`
	Query          string  `json:"query,omitempty"`
	ManagedBy      *string `json:"managedBy,omitempty"`
}

type dimensionEnvelope struct {
	Dimension *Dimension `json:"dimension"`
}

// CreateDimension creates a dimension.
func (c *Client) CreateDimension(ctx context.Context, in DimensionInput) (*Dimension, error) {
	var out dimensionEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/dimensions", in, &out); err != nil {
		return nil, err
	}
	return out.Dimension, nil
}

// GetDimension fetches a dimension by ID.
func (c *Client) GetDimension(ctx context.Context, id string) (*Dimension, error) {
	var out dimensionEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/dimensions/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.Dimension, nil
}

// UpdateDimension updates a dimension by ID. The GrowthBook API uses POST (not
// PUT) for dimension updates.
func (c *Client) UpdateDimension(ctx context.Context, id string, in DimensionInput) (*Dimension, error) {
	var out dimensionEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/dimensions/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return out.Dimension, nil
}

// DeleteDimension deletes a dimension by ID.
func (c *Client) DeleteDimension(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/dimensions/"+url.PathEscape(id), nil, nil)
}

// ListDimensions returns every dimension, following pagination.
func (c *Client) ListDimensions(ctx context.Context) ([]Dimension, error) {
	return fetchAll(ctx, c, "/dimensions", func(b []byte) ([]Dimension, pagination, error) {
		var page struct {
			Dimensions []Dimension `json:"dimensions"`
			pagination
		}
		if err := unmarshal(b, &page); err != nil {
			return nil, pagination{}, err
		}
		return page.Dimensions, page.pagination, nil
	})
}
