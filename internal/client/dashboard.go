package client

import (
	"context"
	"net/http"
	"net/url"
)

// DashboardUpdateSchedule controls automatic refresh of a general dashboard.
type DashboardUpdateSchedule struct {
	Type  string   `json:"type"`
	Hours *float64 `json:"hours,omitempty"`
	Cron  string   `json:"cron,omitempty"`
}

// DashboardBlockLayout positions a block on the dashboard grid.
type DashboardBlockLayout struct {
	X      int64 `json:"x"`
	Y      int64 `json:"y"`
	W      int64 `json:"w"`
	H      int64 `json:"h"`
	Static *bool `json:"static,omitempty"`
}

// DashboardBlock is a single dashboard block. The GrowthBook API supports many
// block types with type-specific fields; this models the fields common to most
// block types. Send the fields appropriate for the chosen block `type`.
type DashboardBlock struct {
	ID           string                `json:"id,omitempty"`
	UID          string                `json:"uid,omitempty"`
	Type         string                `json:"type"`
	Title        string                `json:"title"`
	Description  string                `json:"description"`
	Content      *string               `json:"content,omitempty"`
	SnapshotID   *string               `json:"snapshotId,omitempty"`
	ExperimentID *string               `json:"experimentId,omitempty"`
	MetricIDs    []string              `json:"metricIds,omitempty"`
	VariationIDs []string              `json:"variationIds,omitempty"`
	Layout       *DashboardBlockLayout `json:"layout,omitempty"`
}

// Dashboard mirrors the GrowthBook Dashboard model.
type Dashboard struct {
	ID                string                   `json:"id,omitempty"`
	UID               string                   `json:"uid,omitempty"`
	Organization      string                   `json:"organization,omitempty"`
	ExperimentID      string                   `json:"experimentId,omitempty"`
	IsDefault         bool                     `json:"isDefault,omitempty"`
	UserID            string                   `json:"userId,omitempty"`
	EditLevel         string                   `json:"editLevel,omitempty"`
	ShareLevel        string                   `json:"shareLevel,omitempty"`
	EnableAutoUpdates bool                     `json:"enableAutoUpdates"`
	UpdateSchedule    *DashboardUpdateSchedule `json:"updateSchedule,omitempty"`
	Title             string                   `json:"title,omitempty"`
	Projects          []string                 `json:"projects,omitempty"`
	DateCreated       string                   `json:"dateCreated,omitempty"`
	DateUpdated       string                   `json:"dateUpdated,omitempty"`
	Blocks            []DashboardBlock         `json:"blocks,omitempty"`
}

// DashboardInput is the create/update request body. Blocks is always sent (an
// empty slice rather than null) because the API requires an array.
type DashboardInput struct {
	Title             string                   `json:"title"`
	EditLevel         string                   `json:"editLevel"`
	ShareLevel        string                   `json:"shareLevel"`
	EnableAutoUpdates bool                     `json:"enableAutoUpdates"`
	UpdateSchedule    *DashboardUpdateSchedule `json:"updateSchedule,omitempty"`
	ExperimentID      *string                  `json:"experimentId,omitempty"`
	Projects          []string                 `json:"projects,omitempty"`
	Blocks            []DashboardBlock         `json:"blocks"`
}

type dashboardEnvelope struct {
	Dashboard *Dashboard `json:"dashboard"`
}

// CreateDashboard creates a dashboard.
func (c *Client) CreateDashboard(ctx context.Context, in DashboardInput) (*Dashboard, error) {
	var out dashboardEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/dashboards", in, &out); err != nil {
		return nil, err
	}
	return out.Dashboard, nil
}

// GetDashboard fetches a dashboard by ID.
func (c *Client) GetDashboard(ctx context.Context, id string) (*Dashboard, error) {
	var out dashboardEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/dashboards/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.Dashboard, nil
}

// UpdateDashboard updates a dashboard by ID (HTTP PUT). The update body does not
// accept experimentId (it is set only at creation and is immutable), so it is
// cleared here to avoid an HTTP 400 from the strict (additionalProperties:false)
// update schema.
func (c *Client) UpdateDashboard(ctx context.Context, id string, in DashboardInput) (*Dashboard, error) {
	in.ExperimentID = nil
	var out dashboardEnvelope
	if err := c.doJSON(ctx, http.MethodPut, "/dashboards/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return out.Dashboard, nil
}

// DeleteDashboard deletes a dashboard by ID.
func (c *Client) DeleteDashboard(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/dashboards/"+url.PathEscape(id), nil, nil)
}
