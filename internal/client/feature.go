package client

import (
	"context"
	"net/http"
	"net/url"
)

// SavedGroupTargeting is reused by feature rules and elsewhere.
type SavedGroupTargeting struct {
	MatchType   string   `json:"matchType"`
	SavedGroups []string `json:"savedGroups"`
}

// RulePrerequisite is a feature-ID + condition gate on a single rule.
type RulePrerequisite struct {
	ID        string `json:"id"`
	Condition string `json:"condition"`
}

// ScheduleRule is a simple time-based on/off toggle for a rule.
type ScheduleRule struct {
	Enabled   bool    `json:"enabled"`
	Timestamp *string `json:"timestamp"`
}

// RuleNamespace partitions experiment traffic.
type RuleNamespace struct {
	Enabled bool      `json:"enabled"`
	Name    string    `json:"name"`
	Range   []float64 `json:"range"`
}

// ExperimentValue is a weighted variation value for an inline experiment rule.
type ExperimentValue struct {
	Value  string  `json:"value"`
	Weight float64 `json:"weight"`
	Name   *string `json:"name,omitempty"`
}

// ExperimentRefVariation maps an experiment variation to a feature value.
type ExperimentRefVariation struct {
	Value       string `json:"value"`
	VariationID string `json:"variationId"`
}

// FeatureRule is the flattened union of GrowthBook's force / rollout /
// experiment / experiment-ref rule types. Only the fields relevant to a given
// `Type` are populated; the rest are omitted from the JSON.
type FeatureRule struct {
	ID                     string                   `json:"id,omitempty"`
	Type                   string                   `json:"type"`
	Description            *string                  `json:"description,omitempty"`
	Enabled                *bool                    `json:"enabled,omitempty"`
	Condition              *string                  `json:"condition,omitempty"`
	SavedGroupTargeting    []SavedGroupTargeting    `json:"savedGroupTargeting,omitempty"`
	Prerequisites          []RulePrerequisite       `json:"prerequisites,omitempty"`
	ScheduleRules          []ScheduleRule           `json:"scheduleRules,omitempty"`
	Sparse                 *bool                    `json:"sparse,omitempty"`
	Value                  *string                  `json:"value,omitempty"`
	Coverage               *float64                 `json:"coverage,omitempty"`
	HashAttribute          *string                  `json:"hashAttribute,omitempty"`
	Seed                   *string                  `json:"seed,omitempty"`
	HashVersion            *int64                   `json:"hashVersion,omitempty"`
	TrackingKey            *string                  `json:"trackingKey,omitempty"`
	FallbackAttribute      *string                  `json:"fallbackAttribute,omitempty"`
	DisableStickyBucketing *bool                    `json:"disableStickyBucketing,omitempty"`
	BucketVersion          *int64                   `json:"bucketVersion,omitempty"`
	MinBucketVersion       *int64                   `json:"minBucketVersion,omitempty"`
	Namespace              *RuleNamespace           `json:"namespace,omitempty"`
	Values                 []ExperimentValue        `json:"values,omitempty"`
	Variations             []ExperimentRefVariation `json:"variations,omitempty"`
	ExperimentID           *string                  `json:"experimentId,omitempty"`
}

// FeatureEnvironment is the per-environment settings of a feature.
type FeatureEnvironment struct {
	Enabled bool          `json:"enabled"`
	Rules   []FeatureRule `json:"rules"`
}

// Feature mirrors the GrowthBook FeatureV1 model.
type Feature struct {
	ID            string                        `json:"id"`
	DateCreated   string                        `json:"dateCreated,omitempty"`
	DateUpdated   string                        `json:"dateUpdated,omitempty"`
	Archived      bool                          `json:"archived"`
	Description   string                        `json:"description"`
	Owner         string                        `json:"owner"`
	Project       string                        `json:"project"`
	ValueType     string                        `json:"valueType"`
	DefaultValue  string                        `json:"defaultValue"`
	Tags          []string                      `json:"tags"`
	Prerequisites []string                      `json:"prerequisites,omitempty"`
	Environments  map[string]FeatureEnvironment `json:"environments"`
}

// FeatureEnvironmentInput is the per-environment create/update payload.
//
// Rules is intentionally not `omitempty`: GrowthBook validates
// `environments.<env>.rules` as a required array and rejects a request where the
// key is missing. Callers populate a non-nil (possibly empty) slice so an
// environment with no targeting rules serializes as `[]` rather than being
// dropped.
type FeatureEnvironmentInput struct {
	Enabled *bool         `json:"enabled,omitempty"`
	Rules   []FeatureRule `json:"rules"`
}

// FeatureCreateInput is the POST /features body.
type FeatureCreateInput struct {
	ID           string                             `json:"id"`
	Archived     *bool                              `json:"archived,omitempty"`
	Description  *string                            `json:"description,omitempty"`
	Owner        *string                            `json:"owner,omitempty"`
	Project      *string                            `json:"project,omitempty"`
	ValueType    string                             `json:"valueType,omitempty"`
	DefaultValue *string                            `json:"defaultValue,omitempty"`
	Tags         []string                           `json:"tags,omitempty"`
	Environments map[string]FeatureEnvironmentInput `json:"environments,omitempty"`
}

// FeatureUpdateInput is the POST /features/{id} body. Both `id` and `valueType`
// are immutable: the update endpoint rejects `valueType` as an unrecognized key,
// so a value-type change is handled as a resource replacement instead.
type FeatureUpdateInput struct {
	Archived     *bool                              `json:"archived,omitempty"`
	Description  *string                            `json:"description,omitempty"`
	Owner        *string                            `json:"owner,omitempty"`
	Project      *string                            `json:"project,omitempty"`
	DefaultValue *string                            `json:"defaultValue,omitempty"`
	Tags         []string                           `json:"tags,omitempty"`
	Environments map[string]FeatureEnvironmentInput `json:"environments,omitempty"`
}

type featureEnvelope struct {
	Feature *Feature `json:"feature"`
}

// CreateFeature creates a feature.
func (c *Client) CreateFeature(ctx context.Context, in FeatureCreateInput) (*Feature, error) {
	var out featureEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/features", in, &out); err != nil {
		return nil, err
	}
	return out.Feature, nil
}

// GetFeature fetches a feature by key.
func (c *Client) GetFeature(ctx context.Context, id string) (*Feature, error) {
	var out featureEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/features/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.Feature, nil
}

// UpdateFeature updates a feature by key. GrowthBook uses POST (not PUT) here.
func (c *Client) UpdateFeature(ctx context.Context, id string, in FeatureUpdateInput) (*Feature, error) {
	var out featureEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/features/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return out.Feature, nil
}

// DeleteFeature deletes a feature by key.
func (c *Client) DeleteFeature(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/features/"+url.PathEscape(id), nil, nil)
}
