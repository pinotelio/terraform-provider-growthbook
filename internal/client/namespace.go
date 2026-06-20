package client

import (
	"context"
	"net/http"
	"net/url"
)

// Namespace mirrors the GrowthBook Namespace model. Namespaces partition
// experiment traffic so mutually exclusive experiments don't overlap.
type Namespace struct {
	ID            string `json:"id"`
	DisplayName   string `json:"displayName"`
	Description   string `json:"description"`
	Status        string `json:"status"`
	Format        string `json:"format"`
	HashAttribute string `json:"hashAttribute,omitempty"`
	Seed          string `json:"seed,omitempty"`
}

// NamespaceCreateInput is the POST /namespaces body. Pointer fields are omitted
// from the JSON when nil so server defaults apply.
type NamespaceCreateInput struct {
	DisplayName   string  `json:"displayName"`
	Description   *string `json:"description,omitempty"`
	Status        *string `json:"status,omitempty"`
	Format        *string `json:"format,omitempty"`
	HashAttribute *string `json:"hashAttribute,omitempty"`
}

// NamespaceUpdateInput is the PUT /namespaces/{id} body. The format is immutable
// after creation and is therefore not accepted here.
type NamespaceUpdateInput struct {
	DisplayName   *string `json:"displayName,omitempty"`
	Description   *string `json:"description,omitempty"`
	Status        *string `json:"status,omitempty"`
	HashAttribute *string `json:"hashAttribute,omitempty"`
}

type namespaceEnvelope struct {
	Namespace *Namespace `json:"namespace"`
}

// CreateNamespace creates a namespace.
func (c *Client) CreateNamespace(ctx context.Context, in NamespaceCreateInput) (*Namespace, error) {
	var out namespaceEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/namespaces", in, &out); err != nil {
		return nil, err
	}
	return out.Namespace, nil
}

// GetNamespace fetches a namespace by ID.
func (c *Client) GetNamespace(ctx context.Context, id string) (*Namespace, error) {
	var out namespaceEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/namespaces/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.Namespace, nil
}

// UpdateNamespace updates a namespace by ID.
func (c *Client) UpdateNamespace(ctx context.Context, id string, in NamespaceUpdateInput) (*Namespace, error) {
	var out namespaceEnvelope
	if err := c.doJSON(ctx, http.MethodPut, "/namespaces/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return out.Namespace, nil
}

// DeleteNamespace deletes a namespace by ID.
func (c *Client) DeleteNamespace(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/namespaces/"+url.PathEscape(id), nil, nil)
}

// ListNamespaces returns every namespace, following pagination.
func (c *Client) ListNamespaces(ctx context.Context) ([]Namespace, error) {
	return fetchAll(ctx, c, "/namespaces", func(b []byte) ([]Namespace, pagination, error) {
		var page struct {
			Namespaces []Namespace `json:"namespaces"`
			pagination
		}
		if err := unmarshal(b, &page); err != nil {
			return nil, pagination{}, err
		}
		return page.Namespaces, page.pagination, nil
	})
}
