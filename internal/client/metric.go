package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// MetricCappingSettings controls how metric outliers are handled.
type MetricCappingSettings struct {
	Type        *string  `json:"type,omitempty"`
	Value       *float64 `json:"value,omitempty"`
	IgnoreZeros *bool    `json:"ignoreZeros,omitempty"`
}

// MetricWindowSettings controls the conversion window for a metric.
type MetricWindowSettings struct {
	Type        string   `json:"type,omitempty"`
	DelayValue  *float64 `json:"delayValue,omitempty"`
	DelayUnit   *string  `json:"delayUnit,omitempty"`
	WindowValue *float64 `json:"windowValue,omitempty"`
	WindowUnit  *string  `json:"windowUnit,omitempty"`
}

// MetricPriorSettings controls the bayesian prior for a metric.
type MetricPriorSettings struct {
	Override bool    `json:"override"`
	Proper   bool    `json:"proper"`
	Mean     float64 `json:"mean"`
	Stddev   float64 `json:"stddev"`
}

// MetricBehavior groups the analysis behavior settings of a metric.
type MetricBehavior struct {
	Goal                 *string                `json:"goal,omitempty"`
	CappingSettings      *MetricCappingSettings `json:"cappingSettings,omitempty"`
	WindowSettings       *MetricWindowSettings  `json:"windowSettings,omitempty"`
	PriorSettings        *MetricPriorSettings   `json:"priorSettings,omitempty"`
	RiskThresholdSuccess *float64               `json:"riskThresholdSuccess,omitempty"`
	RiskThresholdDanger  *float64               `json:"riskThresholdDanger,omitempty"`
	MinPercentChange     *float64               `json:"minPercentChange,omitempty"`
	MaxPercentChange     *float64               `json:"maxPercentChange,omitempty"`
	MinSampleSize        *float64               `json:"minSampleSize,omitempty"`
	TargetMDE            *float64               `json:"targetMDE,omitempty"`
}

// MetricSQL defines a SQL-based metric query.
type MetricSQL struct {
	IdentifierTypes     []string `json:"identifierTypes,omitempty"`
	ConversionSQL       string   `json:"conversionSQL,omitempty"`
	UserAggregationSQL  string   `json:"userAggregationSQL,omitempty"`
	DenominatorMetricID string   `json:"denominatorMetricId,omitempty"`
}

// Metric mirrors the GrowthBook (legacy) Metric model. The sqlBuilder and
// mixpanel query definitions are preserved as raw JSON for lossless round-trips
// but are not surfaced as managed Terraform attributes.
type Metric struct {
	ID           string          `json:"id,omitempty"`
	ManagedBy    string          `json:"managedBy,omitempty"`
	Owner        string          `json:"owner,omitempty"`
	DatasourceID string          `json:"datasourceId,omitempty"`
	Name         string          `json:"name,omitempty"`
	Description  string          `json:"description,omitempty"`
	Type         string          `json:"type,omitempty"`
	Tags         []string        `json:"tags,omitempty"`
	Projects     []string        `json:"projects,omitempty"`
	Archived     bool            `json:"archived,omitempty"`
	Behavior     *MetricBehavior `json:"behavior,omitempty"`
	SQL          *MetricSQL      `json:"sql,omitempty"`
	SQLBuilder   json.RawMessage `json:"sqlBuilder,omitempty"`
	Mixpanel     json.RawMessage `json:"mixpanel,omitempty"`
	DateCreated  string          `json:"dateCreated,omitempty"`
	DateUpdated  string          `json:"dateUpdated,omitempty"`
}

// MetricInput is the create/update request body for a metric. Pointer fields
// are omitted from the JSON when nil so server defaults apply.
type MetricInput struct {
	DatasourceID string          `json:"datasourceId,omitempty"`
	ManagedBy    *string         `json:"managedBy,omitempty"`
	Owner        *string         `json:"owner,omitempty"`
	Name         string          `json:"name,omitempty"`
	Description  *string         `json:"description,omitempty"`
	Type         string          `json:"type,omitempty"`
	Tags         []string        `json:"tags,omitempty"`
	Projects     []string        `json:"projects,omitempty"`
	Archived     *bool           `json:"archived,omitempty"`
	Behavior     *MetricBehavior `json:"behavior,omitempty"`
	SQL          *MetricSQL      `json:"sql,omitempty"`
}

type metricEnvelope struct {
	Metric *Metric `json:"metric"`
}

// CreateMetric creates a metric.
func (c *Client) CreateMetric(ctx context.Context, in MetricInput) (*Metric, error) {
	var out metricEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/metrics", in, &out); err != nil {
		return nil, err
	}
	return out.Metric, nil
}

// GetMetric fetches a metric by ID.
func (c *Client) GetMetric(ctx context.Context, id string) (*Metric, error) {
	var out metricEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/metrics/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.Metric, nil
}

// UpdateMetric updates a metric by ID. The datasource of an existing metric is
// immutable, so it is never sent in the update body.
func (c *Client) UpdateMetric(ctx context.Context, id string, in MetricInput) (*Metric, error) {
	in.DatasourceID = ""
	var out metricEnvelope
	if err := c.doJSON(ctx, http.MethodPut, "/metrics/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return out.Metric, nil
}

// DeleteMetric deletes a metric by ID.
func (c *Client) DeleteMetric(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/metrics/"+url.PathEscape(id), nil, nil)
}

// ListMetrics returns every metric, following pagination.
func (c *Client) ListMetrics(ctx context.Context) ([]Metric, error) {
	return fetchAll(ctx, c, "/metrics", func(b []byte) ([]Metric, pagination, error) {
		var page struct {
			Metrics []Metric `json:"metrics"`
			pagination
		}
		if err := unmarshal(b, &page); err != nil {
			return nil, pagination{}, err
		}
		return page.Metrics, page.pagination, nil
	})
}
