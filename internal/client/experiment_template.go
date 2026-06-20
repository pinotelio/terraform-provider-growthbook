package client

import (
	"context"
	"net/http"
	"net/url"
)

// ExperimentTemplateMetadata is the name/description block of a template.
type ExperimentTemplateMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ExperimentTemplateSavedGroup is a saved-group targeting clause.
type ExperimentTemplateSavedGroup struct {
	Match string   `json:"match"`
	IDs   []string `json:"ids"`
}

// ExperimentTemplatePrerequisite is a feature prerequisite for targeting.
type ExperimentTemplatePrerequisite struct {
	ID        string `json:"id"`
	Condition string `json:"condition"`
}

// ExperimentTemplateTargeting holds the default targeting for the template.
type ExperimentTemplateTargeting struct {
	Coverage      float64                          `json:"coverage"`
	Condition     string                           `json:"condition"`
	SavedGroups   []ExperimentTemplateSavedGroup   `json:"savedGroups,omitempty"`
	Prerequisites []ExperimentTemplatePrerequisite `json:"prerequisites,omitempty"`
}

// ExperimentTemplate mirrors the GrowthBook ExperimentTemplate model.
type ExperimentTemplate struct {
	ID                     string                      `json:"id,omitempty"`
	DateCreated            string                      `json:"dateCreated,omitempty"`
	DateUpdated            string                      `json:"dateUpdated,omitempty"`
	Project                string                      `json:"project,omitempty"`
	Owner                  string                      `json:"owner,omitempty"`
	OwnerEmail             string                      `json:"ownerEmail,omitempty"`
	TemplateMetadata       ExperimentTemplateMetadata  `json:"templateMetadata"`
	Type                   string                      `json:"type"`
	Hypothesis             string                      `json:"hypothesis,omitempty"`
	Description            string                      `json:"description,omitempty"`
	Tags                   []string                    `json:"tags,omitempty"`
	Datasource             string                      `json:"datasource"`
	ExposureQueryID        string                      `json:"exposureQueryId"`
	HashAttribute          string                      `json:"hashAttribute,omitempty"`
	FallbackAttribute      string                      `json:"fallbackAttribute,omitempty"`
	DisableStickyBucketing bool                        `json:"disableStickyBucketing,omitempty"`
	GoalMetrics            []string                    `json:"goalMetrics,omitempty"`
	SecondaryMetrics       []string                    `json:"secondaryMetrics,omitempty"`
	GuardrailMetrics       []string                    `json:"guardrailMetrics,omitempty"`
	ActivationMetric       string                      `json:"activationMetric,omitempty"`
	StatsEngine            string                      `json:"statsEngine"`
	Segment                string                      `json:"segment,omitempty"`
	SkipPartialData        bool                        `json:"skipPartialData,omitempty"`
	Targeting              ExperimentTemplateTargeting `json:"targeting"`
}

// ExperimentTemplateInput is the create/update request body. Optional fields are
// pointers so they are omitted from the JSON when unset and server defaults
// apply.
type ExperimentTemplateInput struct {
	Project                *string                     `json:"project,omitempty"`
	TemplateMetadata       ExperimentTemplateMetadata  `json:"templateMetadata"`
	Type                   string                      `json:"type"`
	Hypothesis             *string                     `json:"hypothesis,omitempty"`
	Description            *string                     `json:"description,omitempty"`
	Tags                   []string                    `json:"tags,omitempty"`
	Datasource             string                      `json:"datasource"`
	ExposureQueryID        string                      `json:"exposureQueryId"`
	HashAttribute          *string                     `json:"hashAttribute,omitempty"`
	FallbackAttribute      *string                     `json:"fallbackAttribute,omitempty"`
	DisableStickyBucketing *bool                       `json:"disableStickyBucketing,omitempty"`
	GoalMetrics            []string                    `json:"goalMetrics,omitempty"`
	SecondaryMetrics       []string                    `json:"secondaryMetrics,omitempty"`
	GuardrailMetrics       []string                    `json:"guardrailMetrics,omitempty"`
	ActivationMetric       *string                     `json:"activationMetric,omitempty"`
	StatsEngine            string                      `json:"statsEngine"`
	Segment                *string                     `json:"segment,omitempty"`
	SkipPartialData        *bool                       `json:"skipPartialData,omitempty"`
	Targeting              ExperimentTemplateTargeting `json:"targeting"`
}

type experimentTemplateEnvelope struct {
	ExperimentTemplate *ExperimentTemplate `json:"experimentTemplate"`
}

// CreateExperimentTemplate creates an experiment template.
func (c *Client) CreateExperimentTemplate(ctx context.Context, in ExperimentTemplateInput) (*ExperimentTemplate, error) {
	var out experimentTemplateEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/experiment-templates", in, &out); err != nil {
		return nil, err
	}
	return out.ExperimentTemplate, nil
}

// GetExperimentTemplate fetches an experiment template by ID.
func (c *Client) GetExperimentTemplate(ctx context.Context, id string) (*ExperimentTemplate, error) {
	var out experimentTemplateEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/experiment-templates/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.ExperimentTemplate, nil
}

// UpdateExperimentTemplate updates an experiment template by ID (HTTP PUT).
func (c *Client) UpdateExperimentTemplate(ctx context.Context, id string, in ExperimentTemplateInput) (*ExperimentTemplate, error) {
	var out experimentTemplateEnvelope
	if err := c.doJSON(ctx, http.MethodPut, "/experiment-templates/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return out.ExperimentTemplate, nil
}

// DeleteExperimentTemplate deletes an experiment template by ID.
func (c *Client) DeleteExperimentTemplate(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/experiment-templates/"+url.PathEscape(id), nil, nil)
}
