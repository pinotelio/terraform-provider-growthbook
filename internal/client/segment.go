package client

import (
	"context"
	"net/http"
	"net/url"
)

// Segment mirrors the GrowthBook Segment model.
type Segment struct {
	ID             string   `json:"id,omitempty"`
	Name           string   `json:"name,omitempty"`
	Owner          string   `json:"owner,omitempty"`
	Description    string   `json:"description,omitempty"`
	DatasourceID   string   `json:"datasourceId,omitempty"`
	IdentifierType string   `json:"identifierType,omitempty"`
	Type           string   `json:"type,omitempty"`
	Query          string   `json:"query,omitempty"`
	FactTableID    string   `json:"factTableId,omitempty"`
	Filters        []string `json:"filters,omitempty"`
	Projects       []string `json:"projects,omitempty"`
	ManagedBy      string   `json:"managedBy,omitempty"`
	DateCreated    string   `json:"dateCreated,omitempty"`
	DateUpdated    string   `json:"dateUpdated,omitempty"`
}

// SegmentInput is the create/update request body for a segment. Pointer fields
// are omitted from the JSON when nil so server defaults apply.
type SegmentInput struct {
	Name           string   `json:"name,omitempty"`
	Owner          *string  `json:"owner,omitempty"`
	Description    *string  `json:"description,omitempty"`
	DatasourceID   string   `json:"datasourceId,omitempty"`
	IdentifierType string   `json:"identifierType,omitempty"`
	Type           string   `json:"type,omitempty"`
	Query          *string  `json:"query,omitempty"`
	FactTableID    *string  `json:"factTableId,omitempty"`
	Filters        []string `json:"filters,omitempty"`
	Projects       []string `json:"projects,omitempty"`
	ManagedBy      *string  `json:"managedBy,omitempty"`
}

type segmentEnvelope struct {
	Segment *Segment `json:"segment"`
}

// CreateSegment creates a segment.
func (c *Client) CreateSegment(ctx context.Context, in SegmentInput) (*Segment, error) {
	var out segmentEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/segments", in, &out); err != nil {
		return nil, err
	}
	return out.Segment, nil
}

// GetSegment fetches a segment by ID.
func (c *Client) GetSegment(ctx context.Context, id string) (*Segment, error) {
	var out segmentEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/segments/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.Segment, nil
}

// UpdateSegment updates a segment by ID. The GrowthBook API uses POST (not PUT)
// for segment updates.
func (c *Client) UpdateSegment(ctx context.Context, id string, in SegmentInput) (*Segment, error) {
	var out segmentEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/segments/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return out.Segment, nil
}

// DeleteSegment deletes a segment by ID.
func (c *Client) DeleteSegment(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/segments/"+url.PathEscape(id), nil, nil)
}

// ListSegments returns every segment, following pagination.
func (c *Client) ListSegments(ctx context.Context) ([]Segment, error) {
	return fetchAll(ctx, c, "/segments", func(b []byte) ([]Segment, pagination, error) {
		var page struct {
			Segments []Segment `json:"segments"`
			pagination
		}
		if err := unmarshal(b, &page); err != nil {
			return nil, pagination{}, err
		}
		return page.Segments, page.pagination, nil
	})
}
