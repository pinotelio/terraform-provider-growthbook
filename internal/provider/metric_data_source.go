package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*metricDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*metricDataSource)(nil)
)

func newMetricDataSource() datasource.DataSource { return &metricDataSource{} }

type metricDataSource struct {
	client *client.Client
}

type metricDataSourceModel struct {
	ID           types.String         `tfsdk:"id"`
	Name         types.String         `tfsdk:"name"`
	Description  types.String         `tfsdk:"description"`
	Owner        types.String         `tfsdk:"owner"`
	Type         types.String         `tfsdk:"type"`
	DatasourceID types.String         `tfsdk:"datasource_id"`
	ManagedBy    types.String         `tfsdk:"managed_by"`
	Projects     types.List           `tfsdk:"projects"`
	Tags         types.List           `tfsdk:"tags"`
	Archived     types.Bool           `tfsdk:"archived"`
	Behavior     *metricBehaviorModel `tfsdk:"behavior"`
	SQL          *metricSQLModel      `tfsdk:"sql"`
	DateCreated  types.String         `tfsdk:"date_created"`
	DateUpdated  types.String         `tfsdk:"date_updated"`
}

func (d *metricDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metric"
}

func metricDataSourceComputedFloat(desc string) schema.Float64Attribute {
	return schema.Float64Attribute{Computed: true, Description: desc}
}

func (d *metricDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook (legacy) metric by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Unique metric identifier (e.g. `met_...`).",
			},
			"name":          schema.StringAttribute{Computed: true, Description: "Metric name."},
			"description":   schema.StringAttribute{Computed: true, Description: "Metric description."},
			"owner":         schema.StringAttribute{Computed: true, Description: "Metric owner."},
			"type":          schema.StringAttribute{Computed: true, Description: "Metric type."},
			"datasource_id": schema.StringAttribute{Computed: true, Description: "Data source ID."},
			"managed_by":    schema.StringAttribute{Computed: true, Description: "Where the metric is managed from."},
			"projects": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Project IDs that can access this metric.",
			},
			"tags": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags applied to the metric.",
			},
			"archived": schema.BoolAttribute{Computed: true, Description: "Whether the metric is archived."},
			"behavior": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Analysis behavior settings for the metric.",
				Attributes: map[string]schema.Attribute{
					"goal": schema.StringAttribute{Computed: true, Description: "Metric goal direction."},
					"capping_settings": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Outlier capping settings.",
						Attributes: map[string]schema.Attribute{
							"type":         schema.StringAttribute{Computed: true, Description: "Capping type."},
							"value":        metricDataSourceComputedFloat("Cap value or percentile."),
							"ignore_zeros": schema.BoolAttribute{Computed: true, Description: "Ignore zeros for percentile capping."},
						},
					},
					"window_settings": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Conversion window settings.",
						Attributes: map[string]schema.Attribute{
							"type":         schema.StringAttribute{Computed: true, Description: "Window type."},
							"delay_value":  metricDataSourceComputedFloat("Conversion delay."),
							"delay_unit":   schema.StringAttribute{Computed: true, Description: "Delay unit."},
							"window_value": metricDataSourceComputedFloat("Window length."),
							"window_unit":  schema.StringAttribute{Computed: true, Description: "Window unit."},
						},
					},
					"prior_settings": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Bayesian prior settings.",
						Attributes: map[string]schema.Attribute{
							"override": schema.BoolAttribute{Computed: true, Description: "Override org defaults."},
							"proper":   schema.BoolAttribute{Computed: true, Description: "Use a proper prior."},
							"mean":     metricDataSourceComputedFloat("Prior mean."),
							"stddev":   metricDataSourceComputedFloat("Prior standard deviation."),
						},
					},
					"risk_threshold_success": metricDataSourceComputedFloat("Deprecated risk success threshold."),
					"risk_threshold_danger":  metricDataSourceComputedFloat("Deprecated risk danger threshold."),
					"min_percent_change":     metricDataSourceComputedFloat("Minimum significant percent change."),
					"max_percent_change":     metricDataSourceComputedFloat("Maximum significant percent change."),
					"min_sample_size":        metricDataSourceComputedFloat("Minimum sample size."),
					"target_mde":             metricDataSourceComputedFloat("Target minimum detectable effect."),
				},
			},
			"sql": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "SQL definition of the metric.",
				Attributes: map[string]schema.Attribute{
					"identifier_types": schema.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "Identifier types this metric supports.",
					},
					"conversion_sql":        schema.StringAttribute{Computed: true, Description: "Conversion SQL query."},
					"user_aggregation_sql":  schema.StringAttribute{Computed: true, Description: "User aggregation SQL."},
					"denominator_metric_id": schema.StringAttribute{Computed: true, Description: "Denominator metric ID."},
				},
			},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
		},
	}
}

func (d *metricDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *metricDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data metricDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	mt, err := d.client.GetMetric(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read metric", err.Error())
		return
	}
	data.Name = types.StringValue(mt.Name)
	data.Description = types.StringValue(mt.Description)
	data.Owner = types.StringValue(mt.Owner)
	data.Type = types.StringValue(mt.Type)
	data.DatasourceID = types.StringValue(mt.DatasourceID)
	data.ManagedBy = types.StringValue(mt.ManagedBy)
	data.Projects = sliceToStringList(mt.Projects)
	data.Tags = sliceToStringList(mt.Tags)
	data.Archived = types.BoolValue(mt.Archived)
	data.Behavior = metricBehaviorFromClient(mt.Behavior)
	data.SQL = metricSQLFromClient(mt.SQL)
	data.DateCreated = types.StringValue(mt.DateCreated)
	data.DateUpdated = types.StringValue(mt.DateUpdated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
