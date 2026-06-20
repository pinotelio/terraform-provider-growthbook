package client

import (
	"context"
	"net/http"
	"net/url"
)

// FactMetricColumnRef describes how a fact metric references a column on a fact
// table. It is used for both the numerator and (for ratio metrics) the
// denominator. AggregateFilterColumn/AggregateFilter only apply to the
// numerator and are omitted when empty.
type FactMetricColumnRef struct {
	FactTableID           string   `json:"factTableId"`
	Column                string   `json:"column,omitempty"`
	Aggregation           string   `json:"aggregation,omitempty"`
	Filters               []string `json:"filters,omitempty"`
	AggregateFilterColumn string   `json:"aggregateFilterColumn,omitempty"`
	AggregateFilter       string   `json:"aggregateFilter,omitempty"`
}

// FactMetricCappingSettings controls how outliers are handled.
type FactMetricCappingSettings struct {
	Type        string   `json:"type"`
	Value       *float64 `json:"value,omitempty"`
	IgnoreZeros *bool    `json:"ignoreZeros,omitempty"`
}

// FactMetricWindowSettings controls the conversion window for the metric.
type FactMetricWindowSettings struct {
	Type        string   `json:"type"`
	DelayValue  *float64 `json:"delayValue,omitempty"`
	DelayUnit   string   `json:"delayUnit,omitempty"`
	WindowValue *float64 `json:"windowValue,omitempty"`
	WindowUnit  string   `json:"windowUnit,omitempty"`
}

// FactMetricPriorSettings controls the bayesian prior for the metric.
type FactMetricPriorSettings struct {
	Override bool    `json:"override"`
	Proper   bool    `json:"proper"`
	Mean     float64 `json:"mean"`
	Stddev   float64 `json:"stddev"`
}

// FactMetricRegressionAdjustmentSettings controls CUPED settings for the metric.
type FactMetricRegressionAdjustmentSettings struct {
	Override bool     `json:"override"`
	Enabled  *bool    `json:"enabled,omitempty"`
	Days     *float64 `json:"days,omitempty"`
}

// FactMetricQuantileSettings controls settings for quantile metrics.
type FactMetricQuantileSettings struct {
	Type                     string  `json:"type"`
	IgnoreZeros              bool    `json:"ignoreZeros"`
	Quantile                 float64 `json:"quantile"`
	QuantileEventCountColumn string  `json:"quantileEventCountColumn,omitempty"`
}

// FactMetric mirrors the GrowthBook FactMetric model.
type FactMetric struct {
	ID                           string                                  `json:"id,omitempty"`
	Name                         string                                  `json:"name"`
	Description                  string                                  `json:"description,omitempty"`
	Owner                        string                                  `json:"owner,omitempty"`
	Projects                     []string                                `json:"projects,omitempty"`
	Tags                         []string                                `json:"tags,omitempty"`
	Datasource                   string                                  `json:"datasource,omitempty"`
	MetricType                   string                                  `json:"metricType,omitempty"`
	Numerator                    *FactMetricColumnRef                    `json:"numerator,omitempty"`
	Denominator                  *FactMetricColumnRef                    `json:"denominator,omitempty"`
	Inverse                      bool                                    `json:"inverse,omitempty"`
	QuantileSettings             *FactMetricQuantileSettings             `json:"quantileSettings,omitempty"`
	CappingSettings              *FactMetricCappingSettings              `json:"cappingSettings,omitempty"`
	WindowSettings               *FactMetricWindowSettings               `json:"windowSettings,omitempty"`
	PriorSettings                *FactMetricPriorSettings                `json:"priorSettings,omitempty"`
	RegressionAdjustmentSettings *FactMetricRegressionAdjustmentSettings `json:"regressionAdjustmentSettings,omitempty"`
	DisplayAsPercentage          bool                                    `json:"displayAsPercentage,omitempty"`
	MinPercentChange             float64                                 `json:"minPercentChange,omitempty"`
	MaxPercentChange             float64                                 `json:"maxPercentChange,omitempty"`
	MinSampleSize                float64                                 `json:"minSampleSize,omitempty"`
	TargetMDE                    float64                                 `json:"targetMDE,omitempty"`
	ManagedBy                    string                                  `json:"managedBy,omitempty"`
	MetricAutoSlices             []string                                `json:"metricAutoSlices,omitempty"`
	Archived                     bool                                    `json:"archived,omitempty"`
	DateCreated                  string                                  `json:"dateCreated,omitempty"`
	DateUpdated                  string                                  `json:"dateUpdated,omitempty"`
}

// FactMetricInput is the create/update request body for a fact metric. Pointer
// fields are omitted when nil so server defaults apply. The datasource is
// derived from the numerator's fact table and is not settable. Archived is only
// accepted on update; callers must leave it nil on create.
type FactMetricInput struct {
	Name                         string                                  `json:"name,omitempty"`
	Description                  *string                                 `json:"description,omitempty"`
	Owner                        *string                                 `json:"owner,omitempty"`
	Projects                     []string                                `json:"projects,omitempty"`
	Tags                         []string                                `json:"tags,omitempty"`
	MetricType                   string                                  `json:"metricType,omitempty"`
	Numerator                    *FactMetricColumnRef                    `json:"numerator,omitempty"`
	Denominator                  *FactMetricColumnRef                    `json:"denominator,omitempty"`
	Inverse                      *bool                                   `json:"inverse,omitempty"`
	QuantileSettings             *FactMetricQuantileSettings             `json:"quantileSettings,omitempty"`
	CappingSettings              *FactMetricCappingSettings              `json:"cappingSettings,omitempty"`
	WindowSettings               *FactMetricWindowSettings               `json:"windowSettings,omitempty"`
	PriorSettings                *FactMetricPriorSettings                `json:"priorSettings,omitempty"`
	RegressionAdjustmentSettings *FactMetricRegressionAdjustmentSettings `json:"regressionAdjustmentSettings,omitempty"`
	DisplayAsPercentage          *bool                                   `json:"displayAsPercentage,omitempty"`
	MinPercentChange             *float64                                `json:"minPercentChange,omitempty"`
	MaxPercentChange             *float64                                `json:"maxPercentChange,omitempty"`
	MinSampleSize                *float64                                `json:"minSampleSize,omitempty"`
	TargetMDE                    *float64                                `json:"targetMDE,omitempty"`
	ManagedBy                    *string                                 `json:"managedBy,omitempty"`
	MetricAutoSlices             []string                                `json:"metricAutoSlices,omitempty"`
	Archived                     *bool                                   `json:"archived,omitempty"`
}

type factMetricEnvelope struct {
	FactMetric *FactMetric `json:"factMetric"`
}

// CreateFactMetric creates a fact metric.
func (c *Client) CreateFactMetric(ctx context.Context, in FactMetricInput) (*FactMetric, error) {
	var out factMetricEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/fact-metrics", in, &out); err != nil {
		return nil, err
	}
	return out.FactMetric, nil
}

// GetFactMetric fetches a fact metric by ID.
func (c *Client) GetFactMetric(ctx context.Context, id string) (*FactMetric, error) {
	var out factMetricEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/fact-metrics/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.FactMetric, nil
}

// UpdateFactMetric updates a fact metric by ID. The GrowthBook API uses POST for
// fact metric updates.
func (c *Client) UpdateFactMetric(ctx context.Context, id string, in FactMetricInput) (*FactMetric, error) {
	var out factMetricEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/fact-metrics/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return out.FactMetric, nil
}

// DeleteFactMetric deletes a fact metric by ID.
func (c *Client) DeleteFactMetric(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/fact-metrics/"+url.PathEscape(id), nil, nil)
}
