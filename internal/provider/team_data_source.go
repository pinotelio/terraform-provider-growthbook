package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*teamDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*teamDataSource)(nil)
)

func newTeamDataSource() datasource.DataSource { return &teamDataSource{} }

type teamDataSource struct {
	client *client.Client
}

type teamDataSourceModel struct {
	ID                       types.String           `tfsdk:"id"`
	Name                     types.String           `tfsdk:"name"`
	Description              types.String           `tfsdk:"description"`
	Role                     types.String           `tfsdk:"role"`
	LimitAccessByEnvironment types.Bool             `tfsdk:"limit_access_by_environment"`
	Environments             types.List             `tfsdk:"environments"`
	ProjectRoles             []teamProjectRoleModel `tfsdk:"project_roles"`
	Members                  types.List             `tfsdk:"members"`
	DefaultProject           types.String           `tfsdk:"default_project"`
	CreatedBy                types.String           `tfsdk:"created_by"`
	ManagedByIdp             types.Bool             `tfsdk:"managed_by_idp"`
	DateCreated              types.String           `tfsdk:"date_created"`
	DateUpdated              types.String           `tfsdk:"date_updated"`
}

func (d *teamDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (d *teamDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook team by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Unique team identifier.",
			},
			"name":        schema.StringAttribute{Computed: true, Description: "Team name."},
			"description": schema.StringAttribute{Computed: true, Description: "Team description."},
			"role":        schema.StringAttribute{Computed: true, Description: "Global role for team members."},
			"limit_access_by_environment": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the global role is limited to specific environments.",
			},
			"environments": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Environments the global role is limited to.",
			},
			"project_roles": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "Per-project role overrides for this team.",
				NestedObject: teamProjectRoleDataSourceSchema(),
			},
			"members": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "User IDs that belong to this team.",
			},
			"default_project": schema.StringAttribute{Computed: true, Description: "Default project ID for team members."},
			"created_by":      schema.StringAttribute{Computed: true, Description: "User ID that created the team."},
			"managed_by_idp":  schema.BoolAttribute{Computed: true, Description: "Whether membership is managed by an IdP."},
			"date_created":    schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated":    schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

func teamProjectRoleDataSourceSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"project":                     schema.StringAttribute{Computed: true, Description: "Project ID the role override applies to."},
			"role":                        schema.StringAttribute{Computed: true, Description: "Role granted within the project."},
			"limit_access_by_environment": schema.BoolAttribute{Computed: true, Description: "Whether the override is limited to specific environments."},
			"environments": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Environments the override is limited to.",
			},
			"teams": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Nested team IDs associated with this project role.",
			},
		},
	}
}

func (d *teamDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *teamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data teamDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := d.client.GetTeam(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read team", err.Error())
		return
	}

	data.Name = types.StringValue(team.Name)
	data.Description = types.StringValue(team.Description)
	data.Role = types.StringValue(team.Role)
	data.LimitAccessByEnvironment = types.BoolValue(team.LimitAccessByEnvironment)
	data.Environments = sliceToStringList(team.Environments)
	data.ProjectRoles = flattenTeamProjectRoles(team.ProjectRoles)
	data.Members = sliceToStringList(team.Members)
	data.DefaultProject = types.StringValue(team.DefaultProject)
	data.CreatedBy = types.StringValue(team.CreatedBy)
	data.ManagedByIdp = types.BoolValue(team.ManagedByIdp)
	data.DateCreated = types.StringValue(team.DateCreated)
	data.DateUpdated = types.StringValue(team.DateUpdated)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
