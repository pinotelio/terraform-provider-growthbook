package client

import (
	"context"
	"net/http"
	"net/url"
)

// RampScheduleSavedGroup is a saved-group targeting clause used in step/end patches.
type RampScheduleSavedGroup struct {
	Match string   `json:"match"`
	IDs   []string `json:"ids"`
}

// RampSchedulePrerequisite is a feature prerequisite used in step/end patches.
type RampSchedulePrerequisite struct {
	ID        string `json:"id"`
	Condition string `json:"condition"`
}

// RampSchedulePatch describes the feature-rule mutation applied at a step.
type RampSchedulePatch struct {
	RuleID          string                     `json:"ruleId"`
	Coverage        *float64                   `json:"coverage,omitempty"`
	Condition       *string                    `json:"condition,omitempty"`
	SavedGroups     []RampScheduleSavedGroup   `json:"savedGroups,omitempty"`
	Prerequisites   []RampSchedulePrerequisite `json:"prerequisites,omitempty"`
	AllEnvironments *bool                      `json:"allEnvironments,omitempty"`
	Environments    []string                   `json:"environments,omitempty"`
	Enabled         *bool                      `json:"enabled,omitempty"`
}

// RampScheduleAction targets a feature rule with a patch.
type RampScheduleAction struct {
	TargetType string            `json:"targetType"`
	TargetID   string            `json:"targetId"`
	Patch      RampSchedulePatch `json:"patch"`
}

// RampScheduleHoldConditions gates advancement to the next step.
type RampScheduleHoldConditions struct {
	MinSampleSize    *int64 `json:"minSampleSize,omitempty"`
	RequiresApproval *bool  `json:"requiresApproval,omitempty"`
}

// RampScheduleStep is one stage of a ramp schedule.
type RampScheduleStep struct {
	// Interval is required but nullable; no omitempty so an explicit null is sent.
	Interval       *float64                    `json:"interval"`
	ApprovalNotes  *string                     `json:"approvalNotes,omitempty"`
	Monitored      *bool                       `json:"monitored,omitempty"`
	HoldConditions *RampScheduleHoldConditions `json:"holdConditions,omitempty"`
	Actions        []RampScheduleAction        `json:"actions"`
}

// RampScheduleEndPatch is applied when the schedule completes.
type RampScheduleEndPatch struct {
	Coverage        *float64                   `json:"coverage,omitempty"`
	Condition       *string                    `json:"condition,omitempty"`
	SavedGroups     []RampScheduleSavedGroup   `json:"savedGroups,omitempty"`
	Prerequisites   []RampSchedulePrerequisite `json:"prerequisites,omitempty"`
	AllEnvironments *bool                      `json:"allEnvironments,omitempty"`
	Environments    []string                   `json:"environments,omitempty"`
}

// RampScheduleMonitoringConfig configures automated monitoring of the rollout.
type RampScheduleMonitoringConfig struct {
	DatasourceID              string   `json:"datasourceId"`
	ExposureQueryID           string   `json:"exposureQueryId"`
	GuardrailMetricIDs        []string `json:"guardrailMetricIds"`
	SignalMetricIDs           []string `json:"signalMetricIds,omitempty"`
	UpdateScheduleMinutes     *float64 `json:"updateScheduleMinutes,omitempty"`
	MonitoringMode            string   `json:"monitoringMode,omitempty"`
	AutoUpdate                *bool    `json:"autoUpdate,omitempty"`
	SRMAction                 string   `json:"srmAction,omitempty"`
	NoTrafficAction           string   `json:"noTrafficAction,omitempty"`
	NoTrafficGracePeriodHours *float64 `json:"noTrafficGracePeriodHours,omitempty"`
	MultipleExposureAction    string   `json:"multipleExposureAction,omitempty"`
}

// RampScheduleLockdownConfig controls whether the schedule is locked.
type RampScheduleLockdownConfig struct {
	Mode string `json:"mode"`
}

// RampScheduleTemplate mirrors the GrowthBook RampScheduleTemplate model.
type RampScheduleTemplate struct {
	ID               string                        `json:"id,omitempty"`
	DateCreated      string                        `json:"dateCreated,omitempty"`
	DateUpdated      string                        `json:"dateUpdated,omitempty"`
	Name             string                        `json:"name"`
	Steps            []RampScheduleStep            `json:"steps"`
	EndPatch         *RampScheduleEndPatch         `json:"endPatch,omitempty"`
	Official         *bool                         `json:"official,omitempty"`
	MonitoringConfig *RampScheduleMonitoringConfig `json:"monitoringConfig,omitempty"`
	LockdownConfig   *RampScheduleLockdownConfig   `json:"lockdownConfig,omitempty"`
	Order            *float64                      `json:"order,omitempty"`
}

// RampScheduleTemplateInput is the create/update request body.
type RampScheduleTemplateInput struct {
	Name             string                        `json:"name"`
	Steps            []RampScheduleStep            `json:"steps"`
	EndPatch         *RampScheduleEndPatch         `json:"endPatch,omitempty"`
	Official         *bool                         `json:"official,omitempty"`
	MonitoringConfig *RampScheduleMonitoringConfig `json:"monitoringConfig,omitempty"`
	LockdownConfig   *RampScheduleLockdownConfig   `json:"lockdownConfig,omitempty"`
	Order            *float64                      `json:"order,omitempty"`
}

type rampScheduleTemplateEnvelope struct {
	RampScheduleTemplate *RampScheduleTemplate `json:"rampScheduleTemplate"`
}

// CreateRampScheduleTemplate creates a ramp schedule template.
func (c *Client) CreateRampScheduleTemplate(ctx context.Context, in RampScheduleTemplateInput) (*RampScheduleTemplate, error) {
	var out rampScheduleTemplateEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/ramp-schedule-templates", in, &out); err != nil {
		return nil, err
	}
	return out.RampScheduleTemplate, nil
}

// GetRampScheduleTemplate fetches a ramp schedule template by ID.
func (c *Client) GetRampScheduleTemplate(ctx context.Context, id string) (*RampScheduleTemplate, error) {
	var out rampScheduleTemplateEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/ramp-schedule-templates/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.RampScheduleTemplate, nil
}

// UpdateRampScheduleTemplate updates a ramp schedule template by ID (HTTP PUT).
func (c *Client) UpdateRampScheduleTemplate(ctx context.Context, id string, in RampScheduleTemplateInput) (*RampScheduleTemplate, error) {
	var out rampScheduleTemplateEnvelope
	if err := c.doJSON(ctx, http.MethodPut, "/ramp-schedule-templates/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return out.RampScheduleTemplate, nil
}

// DeleteRampScheduleTemplate deletes a ramp schedule template by ID.
func (c *Client) DeleteRampScheduleTemplate(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/ramp-schedule-templates/"+url.PathEscape(id), nil, nil)
}
