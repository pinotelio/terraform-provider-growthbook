package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*environmentsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*environmentsDataSource)(nil)
)

func newEnvironmentsDataSource() datasource.DataSource { return &environmentsDataSource{} }

type environmentsDataSource struct {
	client *client.Client
}

type environmentsDataSourceModel struct {
	Environments []environmentListItemModel `tfsdk:"environments"`
}

type environmentListItemModel struct {
	ID           types.String `tfsdk:"id"`
	Description  types.String `tfsdk:"description"`
	ToggleOnList types.Bool   `tfsdk:"toggle_on_list"`
	DefaultState types.Bool   `tfsdk:"default_state"`
	Projects     types.List   `tfsdk:"projects"`
}

func (d *environmentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environments"
}

func (d *environmentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Read all GrowthBook environments.",
		Attributes: map[string]schema.Attribute{
			"environments": schema.ListNestedAttribute{
				Computed:    true,
				Description: "All environments in the organization.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":             schema.StringAttribute{Computed: true, Description: "Environment ID/name."},
						"description":    schema.StringAttribute{Computed: true, Description: "Environment description."},
						"toggle_on_list": schema.BoolAttribute{Computed: true, Description: "Whether shown on the feature list page."},
						"default_state":  schema.BoolAttribute{Computed: true, Description: "Default enabled state for new features."},
						"projects": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Project IDs this environment is scoped to.",
						},
					},
				},
			},
		},
	}
}

func (d *environmentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *environmentsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	envs, err := d.client.ListEnvironments(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list environments", err.Error())
		return
	}
	var data environmentsDataSourceModel
	for _, e := range envs {
		data.Environments = append(data.Environments, environmentListItemModel{
			ID:           types.StringValue(e.ID),
			Description:  types.StringValue(e.Description),
			ToggleOnList: types.BoolValue(e.ToggleOnList),
			DefaultState: types.BoolValue(e.DefaultState),
			Projects:     sliceToStringList(e.Projects),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
