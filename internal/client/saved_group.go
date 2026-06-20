package client

import (
	"context"
	"net/http"
	"net/url"
)

// SavedGroup mirrors the GrowthBook SavedGroup model. A saved group is either a
// `list` group (an attribute key + a list of values) or a `condition` group (a
// JSON-encoded targeting condition).
type SavedGroup struct {
	ID                string   `json:"id"`
	Type              string   `json:"type"`
	Name              string   `json:"name"`
	Owner             string   `json:"owner,omitempty"`
	Condition         string   `json:"condition,omitempty"`
	AttributeKey      string   `json:"attributeKey,omitempty"`
	Values            []string `json:"values,omitempty"`
	Description       string   `json:"description,omitempty"`
	Projects          []string `json:"projects,omitempty"`
	Archived          bool     `json:"archived"`
	UseEmptyListGroup bool     `json:"useEmptyListGroup"`
	DateCreated       string   `json:"dateCreated,omitempty"`
	DateUpdated       string   `json:"dateUpdated,omitempty"`
}

// SavedGroupCreateInput is the POST /saved-groups body. Pointer fields are
// omitted from the JSON when nil so server defaults apply.
type SavedGroupCreateInput struct {
	Name         string   `json:"name"`
	Type         *string  `json:"type,omitempty"`
	Condition    *string  `json:"condition,omitempty"`
	AttributeKey *string  `json:"attributeKey,omitempty"`
	Values       []string `json:"values,omitempty"`
	Owner        *string  `json:"owner,omitempty"`
	Projects     []string `json:"projects,omitempty"`
}

// SavedGroupUpdateInput is the POST /saved-groups/{id} body (the API uses POST,
// not PUT, for updates). The group's `type` and `attributeKey` are immutable
// and therefore not accepted here.
type SavedGroupUpdateInput struct {
	Name      *string  `json:"name,omitempty"`
	Condition *string  `json:"condition,omitempty"`
	Values    []string `json:"values,omitempty"`
	Owner     *string  `json:"owner,omitempty"`
	Projects  []string `json:"projects,omitempty"`
}

type savedGroupEnvelope struct {
	SavedGroup *SavedGroup `json:"savedGroup"`
}

// CreateSavedGroup creates a saved group.
func (c *Client) CreateSavedGroup(ctx context.Context, in SavedGroupCreateInput) (*SavedGroup, error) {
	var out savedGroupEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/saved-groups", in, &out); err != nil {
		return nil, err
	}
	return out.SavedGroup, nil
}

// GetSavedGroup fetches a saved group by ID.
func (c *Client) GetSavedGroup(ctx context.Context, id string) (*SavedGroup, error) {
	var out savedGroupEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/saved-groups/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.SavedGroup, nil
}

// UpdateSavedGroup partially updates a saved group by ID. The GrowthBook API
// uses POST (not PUT) for saved group updates.
func (c *Client) UpdateSavedGroup(ctx context.Context, id string, in SavedGroupUpdateInput) (*SavedGroup, error) {
	var out savedGroupEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/saved-groups/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return out.SavedGroup, nil
}

// DeleteSavedGroup deletes a saved group by ID.
func (c *Client) DeleteSavedGroup(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/saved-groups/"+url.PathEscape(id), nil, nil)
}
