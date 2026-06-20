package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*projectsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*projectsDataSource)(nil)
)

func newProjectsDataSource() datasource.DataSource { return &projectsDataSource{} }

type projectsDataSource struct {
	client *client.Client
}

type projectsDataSourceModel struct {
	Projects []projectListItemModel `tfsdk:"projects"`
}

type projectListItemModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	StatsEngine types.String `tfsdk:"stats_engine"`
}

func (d *projectsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

func (d *projectsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Read all GrowthBook projects.",
		Attributes: map[string]schema.Attribute{
			"projects": schema.ListNestedAttribute{
				Computed:    true,
				Description: "All projects in the organization.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true, Description: "Project ID."},
						"name":         schema.StringAttribute{Computed: true, Description: "Project name."},
						"description":  schema.StringAttribute{Computed: true, Description: "Project description."},
						"stats_engine": schema.StringAttribute{Computed: true, Description: "Default statistics engine."},
					},
				},
			},
		},
	}
}

func (d *projectsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *projectsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	projects, err := d.client.ListProjects(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list projects", err.Error())
		return
	}
	var data projectsDataSourceModel
	for _, p := range projects {
		data.Projects = append(data.Projects, projectListItemModel{
			ID:          types.StringValue(p.ID),
			Name:        types.StringValue(p.Name),
			Description: types.StringValue(p.Description),
			StatsEngine: types.StringValue(p.Settings.StatsEngine),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
