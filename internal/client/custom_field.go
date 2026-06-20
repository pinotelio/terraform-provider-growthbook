package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// CustomField mirrors the GrowthBook CustomField model. DefaultValue is kept as
// raw JSON because the API accepts several shapes (string, number, boolean,
// date, or arrays of those).
type CustomField struct {
	ID           string          `json:"id,omitempty"`
	Name         string          `json:"name"`
	Description  string          `json:"description,omitempty"`
	Placeholder  string          `json:"placeholder,omitempty"`
	DefaultValue json.RawMessage `json:"defaultValue,omitempty"`
	Type         string          `json:"type"`
	Values       string          `json:"values,omitempty"`
	Required     bool            `json:"required"`
	Creator      string          `json:"creator,omitempty"`
	Projects     []string        `json:"projects,omitempty"`
	Sections     []string        `json:"sections"`
	Active       bool            `json:"active"`
	DateCreated  string          `json:"dateCreated,omitempty"`
	DateUpdated  string          `json:"dateUpdated,omitempty"`
}

// CustomFieldCreateInput is the create request body. The API requires the
// caller to supply the unique key (id), name, type, required, and sections.
type CustomFieldCreateInput struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Description  *string         `json:"description,omitempty"`
	Placeholder  *string         `json:"placeholder,omitempty"`
	DefaultValue json.RawMessage `json:"defaultValue,omitempty"`
	Type         string          `json:"type"`
	Values       *string         `json:"values,omitempty"`
	Required     bool            `json:"required"`
	Projects     []string        `json:"projects,omitempty"`
	Sections     []string        `json:"sections"`
}

// CustomFieldUpdateInput is the update request body. The id and type cannot be
// changed and are therefore not part of the update payload.
type CustomFieldUpdateInput struct {
	Name         string          `json:"name,omitempty"`
	Description  *string         `json:"description,omitempty"`
	Placeholder  *string         `json:"placeholder,omitempty"`
	DefaultValue json.RawMessage `json:"defaultValue,omitempty"`
	Values       *string         `json:"values,omitempty"`
	Required     *bool           `json:"required,omitempty"`
	Projects     []string        `json:"projects,omitempty"`
	Sections     []string        `json:"sections,omitempty"`
	Active       *bool           `json:"active,omitempty"`
}

type customFieldEnvelope struct {
	CustomField *CustomField `json:"customField"`
}

// CreateCustomField creates a custom field.
func (c *Client) CreateCustomField(ctx context.Context, in CustomFieldCreateInput) (*CustomField, error) {
	var out customFieldEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/custom-fields", in, &out); err != nil {
		return nil, err
	}
	return out.CustomField, nil
}

// GetCustomField fetches a custom field by ID.
func (c *Client) GetCustomField(ctx context.Context, id string) (*CustomField, error) {
	var out customFieldEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/custom-fields/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.CustomField, nil
}

// UpdateCustomField updates a custom field by ID (PUT).
func (c *Client) UpdateCustomField(ctx context.Context, id string, in CustomFieldUpdateInput) (*CustomField, error) {
	var out customFieldEnvelope
	if err := c.doJSON(ctx, http.MethodPut, "/custom-fields/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return out.CustomField, nil
}

// DeleteCustomField deletes a custom field by ID.
func (c *Client) DeleteCustomField(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/custom-fields/"+url.PathEscape(id), nil, nil)
}
