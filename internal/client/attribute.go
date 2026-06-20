package client

import (
	"context"
	"net/http"
	"net/url"
)

// Attribute mirrors the GrowthBook Attribute model. A targeting attribute is
// identified by its `property`, which doubles as its primary key.
type Attribute struct {
	Property      string   `json:"property"`
	Datatype      string   `json:"datatype"`
	Description   string   `json:"description,omitempty"`
	HashAttribute bool     `json:"hashAttribute"`
	Archived      bool     `json:"archived"`
	Enum          string   `json:"enum,omitempty"`
	Format        string   `json:"format,omitempty"`
	Projects      []string `json:"projects,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

// AttributeCreateInput is the POST /attributes body. Pointer fields are omitted
// from the JSON when nil so server defaults apply.
type AttributeCreateInput struct {
	Property      string   `json:"property"`
	Datatype      string   `json:"datatype"`
	Description   *string  `json:"description,omitempty"`
	Archived      *bool    `json:"archived,omitempty"`
	HashAttribute *bool    `json:"hashAttribute,omitempty"`
	Enum          *string  `json:"enum,omitempty"`
	Format        *string  `json:"format,omitempty"`
	Projects      []string `json:"projects,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

// AttributeUpdateInput is the PUT /attributes/{property} body. The property
// itself is immutable and supplied via the path.
type AttributeUpdateInput struct {
	Datatype      *string  `json:"datatype,omitempty"`
	Description   *string  `json:"description,omitempty"`
	Archived      *bool    `json:"archived,omitempty"`
	HashAttribute *bool    `json:"hashAttribute,omitempty"`
	Enum          *string  `json:"enum,omitempty"`
	Format        *string  `json:"format,omitempty"`
	Projects      []string `json:"projects,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

type attributeEnvelope struct {
	Attribute *Attribute `json:"attribute"`
}

// CreateAttribute creates a targeting attribute.
func (c *Client) CreateAttribute(ctx context.Context, in AttributeCreateInput) (*Attribute, error) {
	var out attributeEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/attributes", in, &out); err != nil {
		return nil, err
	}
	return out.Attribute, nil
}

// UpdateAttribute updates a targeting attribute by property.
func (c *Client) UpdateAttribute(ctx context.Context, property string, in AttributeUpdateInput) (*Attribute, error) {
	var out attributeEnvelope
	if err := c.doJSON(ctx, http.MethodPut, "/attributes/"+url.PathEscape(property), in, &out); err != nil {
		return nil, err
	}
	return out.Attribute, nil
}

// DeleteAttribute deletes a targeting attribute by property.
func (c *Client) DeleteAttribute(ctx context.Context, property string) error {
	return c.doJSON(ctx, http.MethodDelete, "/attributes/"+url.PathEscape(property), nil, nil)
}

// ListAttributes returns all targeting attributes. The endpoint is not
// paginated.
func (c *Client) ListAttributes(ctx context.Context) ([]Attribute, error) {
	var out struct {
		Attributes []Attribute `json:"attributes"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/attributes", nil, &out); err != nil {
		return nil, err
	}
	return out.Attributes, nil
}

// GetAttribute returns a single attribute by property. GrowthBook has no
// get-by-id endpoint for attributes, so this filters the list.
func (c *Client) GetAttribute(ctx context.Context, property string) (*Attribute, error) {
	attrs, err := c.ListAttributes(ctx)
	if err != nil {
		return nil, err
	}
	for i := range attrs {
		if attrs[i].Property == property {
			return &attrs[i], nil
		}
	}
	return nil, ErrNotFound
}
