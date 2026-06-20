package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// Archetype mirrors the GrowthBook Archetype model. Attributes is an arbitrary
// JSON object and is kept as raw JSON so the provider can round-trip it as a
// string without imposing a fixed shape.
type Archetype struct {
	ID           string          `json:"id,omitempty"`
	Name         string          `json:"name"`
	Description  string          `json:"description,omitempty"`
	Owner        string          `json:"owner,omitempty"`
	OwnerEmail   string          `json:"ownerEmail,omitempty"`
	IsPublic     bool            `json:"isPublic"`
	Attributes   json.RawMessage `json:"attributes,omitempty"`
	Projects     []string        `json:"projects,omitempty"`
	Environments []string        `json:"environments,omitempty"`
	DateCreated  string          `json:"dateCreated,omitempty"`
	DateUpdated  string          `json:"dateUpdated,omitempty"`
}

// ArchetypeInput is the create/update request body for an archetype.
type ArchetypeInput struct {
	Name         string          `json:"name,omitempty"`
	Description  *string         `json:"description,omitempty"`
	IsPublic     bool            `json:"isPublic"`
	Attributes   json.RawMessage `json:"attributes,omitempty"`
	Projects     []string        `json:"projects,omitempty"`
	Environments []string        `json:"environments,omitempty"`
}

type archetypeEnvelope struct {
	Archetype *Archetype `json:"archetype"`
}

// CreateArchetype creates an archetype.
func (c *Client) CreateArchetype(ctx context.Context, in ArchetypeInput) (*Archetype, error) {
	var out archetypeEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/archetypes", in, &out); err != nil {
		return nil, err
	}
	return out.Archetype, nil
}

// GetArchetype fetches an archetype by ID.
func (c *Client) GetArchetype(ctx context.Context, id string) (*Archetype, error) {
	var out archetypeEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/archetypes/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.Archetype, nil
}

// UpdateArchetype updates an archetype by ID (PUT).
func (c *Client) UpdateArchetype(ctx context.Context, id string, in ArchetypeInput) (*Archetype, error) {
	var out archetypeEnvelope
	if err := c.doJSON(ctx, http.MethodPut, "/archetypes/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return out.Archetype, nil
}

// DeleteArchetype deletes an archetype by ID.
func (c *Client) DeleteArchetype(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/archetypes/"+url.PathEscape(id), nil, nil)
}
