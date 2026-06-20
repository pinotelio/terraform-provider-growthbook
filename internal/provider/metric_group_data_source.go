package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*metricGroupDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*metricGroupDataSource)(nil)
)

func newMetricGroupDataSource() datasource.DataSource { return &metricGroupDataSource{} }

type metricGroupDataSource struct {
	client *client.Client
}

type metricGroupDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Owner       types.String `tfsdk:"owner"`
	Description types.String `tfsdk:"description"`
	Tags        types.List   `tfsdk:"tags"`
	Projects    types.List   `tfsdk:"projects"`
	Metrics     types.List   `tfsdk:"metrics"`
	Datasource  types.String `tfsdk:"datasource"`
	Archived    types.Bool   `tfsdk:"archived"`
	DateCreated types.String `tfsdk:"date_created"`
	DateUpdated types.String `tfsdk:"date_updated"`
}

func (d *metricGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metric_group"
}

func (d *metricGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook metric group by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Unique metric group identifier (e.g. `mg_...`).",
			},
			"name":        schema.StringAttribute{Computed: true, Description: "Metric group name."},
			"owner":       schema.StringAttribute{Computed: true, Description: "Metric group owner."},
			"description": schema.StringAttribute{Computed: true, Description: "Metric group description."},
			"tags": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags applied to the metric group.",
			},
			"projects": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Project IDs the metric group is scoped to.",
			},
			"metrics": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Ordered list of metric IDs in the group.",
			},
			"datasource":   schema.StringAttribute{Computed: true, Description: "Data source ID."},
			"archived":     schema.BoolAttribute{Computed: true, Description: "Whether the metric group is archived."},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
		},
	}
}

func (d *metricGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *metricGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data metricGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	group, err := d.client.GetMetricGroup(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read metric group", err.Error())
		return
	}
	data.Name = types.StringValue(group.Name)
	data.Owner = types.StringValue(group.Owner)
	data.Description = types.StringValue(group.Description)
	data.Tags = sliceToStringList(group.Tags)
	data.Projects = sliceToStringList(group.Projects)
	data.Metrics = sliceToStringList(group.Metrics)
	data.Datasource = types.StringValue(group.Datasource)
	data.Archived = types.BoolValue(group.Archived)
	data.DateCreated = types.StringValue(group.DateCreated)
	data.DateUpdated = types.StringValue(group.DateUpdated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
