package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*environmentDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*environmentDataSource)(nil)
)

func newEnvironmentDataSource() datasource.DataSource { return &environmentDataSource{} }

type environmentDataSource struct {
	client *client.Client
}

type environmentDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Description  types.String `tfsdk:"description"`
	ToggleOnList types.Bool   `tfsdk:"toggle_on_list"`
	DefaultState types.Bool   `tfsdk:"default_state"`
	Projects     types.List   `tfsdk:"projects"`
	Parent       types.String `tfsdk:"parent"`
}

func (d *environmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (d *environmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook environment by ID.",
		Attributes: map[string]schema.Attribute{
			"id":             schema.StringAttribute{Required: true, Description: "The environment identifier/name."},
			"description":    schema.StringAttribute{Computed: true, Description: "Description of the environment."},
			"toggle_on_list": schema.BoolAttribute{Computed: true, Description: "Whether shown on the feature list page."},
			"default_state":  schema.BoolAttribute{Computed: true, Description: "Default enabled state for new features."},
			"projects": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Project IDs this environment is scoped to.",
			},
			"parent": schema.StringAttribute{Computed: true, Description: "Parent environment, if any."},
		},
	}
}

func (d *environmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *environmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data environmentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, err := d.client.GetEnvironment(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read environment", err.Error())
		return
	}

	data.Description = types.StringValue(env.Description)
	data.ToggleOnList = types.BoolValue(env.ToggleOnList)
	data.DefaultState = types.BoolValue(env.DefaultState)
	data.Projects = sliceToStringList(env.Projects)
	if env.Parent == "" {
		data.Parent = types.StringNull()
	} else {
		data.Parent = types.StringValue(env.Parent)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
