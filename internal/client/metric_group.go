package client

import (
	"context"
	"net/http"
	"net/url"
)

// MetricGroup mirrors the GrowthBook MetricGroup model.
type MetricGroup struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`
	Owner       string   `json:"owner,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Projects    []string `json:"projects,omitempty"`
	Metrics     []string `json:"metrics,omitempty"`
	Datasource  string   `json:"datasource,omitempty"`
	Archived    bool     `json:"archived,omitempty"`
	DateCreated string   `json:"dateCreated,omitempty"`
	DateUpdated string   `json:"dateUpdated,omitempty"`
}

// MetricGroupInput is the create/update request body for a metric group.
type MetricGroupInput struct {
	Name        string   `json:"name,omitempty"`
	Owner       *string  `json:"owner,omitempty"`
	Description string   `json:"description"`
	Tags        []string `json:"tags,omitempty"`
	Projects    []string `json:"projects"`
	Metrics     []string `json:"metrics"`
	Datasource  string   `json:"datasource,omitempty"`
	Archived    *bool    `json:"archived,omitempty"`
}

type metricGroupEnvelope struct {
	MetricGroup *MetricGroup `json:"metricGroup"`
}

// CreateMetricGroup creates a metric group.
func (c *Client) CreateMetricGroup(ctx context.Context, in MetricGroupInput) (*MetricGroup, error) {
	var out metricGroupEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/metric-groups", in, &out); err != nil {
		return nil, err
	}
	return out.MetricGroup, nil
}

// GetMetricGroup fetches a metric group by ID.
func (c *Client) GetMetricGroup(ctx context.Context, id string) (*MetricGroup, error) {
	var out metricGroupEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/metric-groups/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.MetricGroup, nil
}

// UpdateMetricGroup updates a metric group by ID.
func (c *Client) UpdateMetricGroup(ctx context.Context, id string, in MetricGroupInput) (*MetricGroup, error) {
	var out metricGroupEnvelope
	if err := c.doJSON(ctx, http.MethodPut, "/metric-groups/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return out.MetricGroup, nil
}

// DeleteMetricGroup deletes a metric group by ID.
func (c *Client) DeleteMetricGroup(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/metric-groups/"+url.PathEscape(id), nil, nil)
}

// ListMetricGroups returns every metric group, following pagination.
func (c *Client) ListMetricGroups(ctx context.Context) ([]MetricGroup, error) {
	return fetchAll(ctx, c, "/metric-groups", func(b []byte) ([]MetricGroup, pagination, error) {
		var page struct {
			MetricGroups []MetricGroup `json:"metricGroups"`
			pagination
		}
		if err := unmarshal(b, &page); err != nil {
			return nil, pagination{}, err
		}
		return page.MetricGroups, page.pagination, nil
	})
}
