package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*factMetricDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*factMetricDataSource)(nil)
)

func newFactMetricDataSource() datasource.DataSource { return &factMetricDataSource{} }

type factMetricDataSource struct {
	client *client.Client
}

func (d *factMetricDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fact_metric"
}

func factMetricDataSourceColumnRefSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed:    true,
		Description: "Reference to a fact table column, including aggregation and filters.",
		Attributes: map[string]schema.Attribute{
			"fact_table_id": schema.StringAttribute{Computed: true, Description: "Fact table ID referenced."},
			"column":        schema.StringAttribute{Computed: true, Description: "Column name or special value."},
			"aggregation":   schema.StringAttribute{Computed: true, Description: "User aggregation of the column."},
			"filters": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Fact table filter IDs applied to this column.",
			},
			"aggregate_filter_column": schema.StringAttribute{Computed: true, Description: "Column used to filter users after aggregation."},
			"aggregate_filter":        schema.StringAttribute{Computed: true, Description: "Comparison applied after aggregation."},
		},
	}
}

func (d *factMetricDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook fact metric by ID.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Required: true, Description: "Unique fact metric identifier."},
			"name":        schema.StringAttribute{Computed: true, Description: "Fact metric name."},
			"description": schema.StringAttribute{Computed: true, Description: "Fact metric description."},
			"owner":       schema.StringAttribute{Computed: true, Description: "Owner userId or email address."},
			"projects": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Associated project IDs.",
			},
			"tags": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags associated with the fact metric.",
			},
			"datasource":  schema.StringAttribute{Computed: true, Description: "Data source ID."},
			"metric_type": schema.StringAttribute{Computed: true, Description: "Metric type."},
			"numerator":   factMetricDataSourceColumnRefSchema(),
			"denominator": factMetricDataSourceColumnRefSchema(),
			"inverse":     schema.BoolAttribute{Computed: true, Description: "Whether a decrease is good."},
			"quantile_settings": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Settings for quantile metrics.",
				Attributes: map[string]schema.Attribute{
					"type":                        schema.StringAttribute{Computed: true, Description: "Quantile over `event` or `unit`."},
					"ignore_zeros":                schema.BoolAttribute{Computed: true, Description: "Ignore zero values."},
					"quantile":                    schema.Float64Attribute{Computed: true, Description: "Quantile value."},
					"quantile_event_count_column": schema.StringAttribute{Computed: true, Description: "Per-row event count column override."},
				},
			},
			"capping_settings": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Controls how outliers are handled.",
				Attributes: map[string]schema.Attribute{
					"type":         schema.StringAttribute{Computed: true, Description: "Capping type."},
					"value":        schema.Float64Attribute{Computed: true, Description: "Absolute value or percentile."},
					"ignore_zeros": schema.BoolAttribute{Computed: true, Description: "Ignore zeros when computing a percentile cap."},
				},
			},
			"window_settings": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Controls the conversion window.",
				Attributes: map[string]schema.Attribute{
					"type":         schema.StringAttribute{Computed: true, Description: "Window type."},
					"delay_value":  schema.Float64Attribute{Computed: true, Description: "Delay after exposure."},
					"delay_unit":   schema.StringAttribute{Computed: true, Description: "Delay unit."},
					"window_value": schema.Float64Attribute{Computed: true, Description: "Window length."},
					"window_unit":  schema.StringAttribute{Computed: true, Description: "Window unit."},
				},
			},
			"prior_settings": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Controls the bayesian prior.",
				Attributes: map[string]schema.Attribute{
					"override": schema.BoolAttribute{Computed: true, Description: "Whether org defaults are overridden."},
					"proper":   schema.BoolAttribute{Computed: true, Description: "Whether a proper prior is used."},
					"mean":     schema.Float64Attribute{Computed: true, Description: "Prior mean."},
					"stddev":   schema.Float64Attribute{Computed: true, Description: "Prior standard deviation."},
				},
			},
			"regression_adjustment_settings": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Controls regression adjustment (CUPED) settings.",
				Attributes: map[string]schema.Attribute{
					"override": schema.BoolAttribute{Computed: true, Description: "Whether org defaults are overridden."},
					"enabled":  schema.BoolAttribute{Computed: true, Description: "Whether regression adjustment is applied."},
					"days":     schema.Float64Attribute{Computed: true, Description: "Pre-exposure days used."},
				},
			},
			"display_as_percentage": schema.BoolAttribute{Computed: true, Description: "Whether means display as a percentage."},
			"min_percent_change":    schema.Float64Attribute{Computed: true, Description: "Minimum significant percent change."},
			"max_percent_change":    schema.Float64Attribute{Computed: true, Description: "Maximum significant percent change."},
			"min_sample_size":       schema.Float64Attribute{Computed: true, Description: "Minimum sample size."},
			"target_mde":            schema.Float64Attribute{Computed: true, Description: "Target minimum detectable effect."},
			"managed_by":            schema.StringAttribute{Computed: true, Description: "Where the metric is managed from."},
			"metric_auto_slices": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Slice column names included in metric analysis.",
			},
			"archived":     schema.BoolAttribute{Computed: true, Description: "Whether the fact metric is archived."},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

func (d *factMetricDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *factMetricDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data factMetricResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fm, err := d.client.GetFactMetric(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read fact metric", err.Error())
		return
	}
	(&factMetricResource{}).apply(&data, fm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
