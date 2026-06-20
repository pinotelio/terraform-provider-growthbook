package client

import (
	"context"
	"net/http"
	"net/url"
)

// FactTableColumn mirrors a column definition on a GrowthBook fact table. The
// set of columns is derived from parsing the fact table SQL; the API only
// allows updating metadata on existing columns, not creating or deleting them.
type FactTableColumn struct {
	Column             string   `json:"column"`
	Datatype           string   `json:"datatype"`
	NumberFormat       string   `json:"numberFormat,omitempty"`
	Name               string   `json:"name,omitempty"`
	Description        string   `json:"description,omitempty"`
	AlwaysInlineFilter bool     `json:"alwaysInlineFilter"`
	Deleted            bool     `json:"deleted"`
	IsAutoSliceColumn  bool     `json:"isAutoSliceColumn"`
	AutoSlices         []string `json:"autoSlices,omitempty"`
	LockedAutoSlices   []string `json:"lockedAutoSlices,omitempty"`
	DateCreated        string   `json:"dateCreated,omitempty"`
	DateUpdated        string   `json:"dateUpdated,omitempty"`
}

// FactTable mirrors the GrowthBook FactTable model.
type FactTable struct {
	ID          string            `json:"id,omitempty"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Owner       string            `json:"owner,omitempty"`
	Projects    []string          `json:"projects,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Datasource  string            `json:"datasource,omitempty"`
	UserIDTypes []string          `json:"userIdTypes,omitempty"`
	SQL         string            `json:"sql,omitempty"`
	EventName   string            `json:"eventName,omitempty"`
	Columns     []FactTableColumn `json:"columns,omitempty"`
	Archived    bool              `json:"archived,omitempty"`
	ManagedBy   string            `json:"managedBy,omitempty"`
	DateCreated string            `json:"dateCreated,omitempty"`
	DateUpdated string            `json:"dateUpdated,omitempty"`
}

// FactTableCreateInput is the POST /fact-tables request body. The datasource is
// only settable on creation.
type FactTableCreateInput struct {
	Name        string   `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Owner       *string  `json:"owner,omitempty"`
	Projects    []string `json:"projects,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Datasource  string   `json:"datasource,omitempty"`
	UserIDTypes []string `json:"userIdTypes,omitempty"`
	SQL         string   `json:"sql,omitempty"`
	EventName   *string  `json:"eventName,omitempty"`
	ManagedBy   *string  `json:"managedBy,omitempty"`
}

// FactTableUpdateInput is the POST /fact-tables/{id} request body. Note the API
// uses POST (not PUT) for updates and does not accept a datasource change.
type FactTableUpdateInput struct {
	Name        string            `json:"name,omitempty"`
	Description *string           `json:"description,omitempty"`
	Owner       *string           `json:"owner,omitempty"`
	Projects    []string          `json:"projects,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	UserIDTypes []string          `json:"userIdTypes,omitempty"`
	SQL         string            `json:"sql,omitempty"`
	EventName   *string           `json:"eventName,omitempty"`
	Columns     []FactTableColumn `json:"columns,omitempty"`
	ManagedBy   *string           `json:"managedBy,omitempty"`
	Archived    *bool             `json:"archived,omitempty"`
}

type factTableEnvelope struct {
	FactTable *FactTable `json:"factTable"`
}

// CreateFactTable creates a fact table.
func (c *Client) CreateFactTable(ctx context.Context, in FactTableCreateInput) (*FactTable, error) {
	var out factTableEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/fact-tables", in, &out); err != nil {
		return nil, err
	}
	return out.FactTable, nil
}

// GetFactTable fetches a fact table by ID.
func (c *Client) GetFactTable(ctx context.Context, id string) (*FactTable, error) {
	var out factTableEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/fact-tables/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.FactTable, nil
}

// UpdateFactTable updates a fact table by ID. The GrowthBook API uses POST for
// fact table updates.
func (c *Client) UpdateFactTable(ctx context.Context, id string, in FactTableUpdateInput) (*FactTable, error) {
	var out factTableEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/fact-tables/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return out.FactTable, nil
}

// DeleteFactTable deletes a fact table by ID.
func (c *Client) DeleteFactTable(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/fact-tables/"+url.PathEscape(id), nil, nil)
}
