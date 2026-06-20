package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*projectDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*projectDataSource)(nil)
)

func newProjectDataSource() datasource.DataSource { return &projectDataSource{} }

type projectDataSource struct {
	client *client.Client
}

type projectDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	StatsEngine types.String `tfsdk:"stats_engine"`
	DateCreated types.String `tfsdk:"date_created"`
	DateUpdated types.String `tfsdk:"date_updated"`
}

func (d *projectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *projectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook project by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Unique project identifier (e.g. `prj_...`).",
			},
			"name":         schema.StringAttribute{Computed: true, Description: "Project name."},
			"description":  schema.StringAttribute{Computed: true, Description: "Project description."},
			"stats_engine": schema.StringAttribute{Computed: true, Description: "Default statistics engine."},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

func (d *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data projectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := d.client.GetProject(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read project", err.Error())
		return
	}

	data.Name = types.StringValue(project.Name)
	data.Description = types.StringValue(project.Description)
	data.StatsEngine = types.StringValue(project.Settings.StatsEngine)
	data.DateCreated = types.StringValue(project.DateCreated)
	data.DateUpdated = types.StringValue(project.DateUpdated)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
