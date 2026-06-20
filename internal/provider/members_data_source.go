package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*membersDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*membersDataSource)(nil)
)

func newMembersDataSource() datasource.DataSource { return &membersDataSource{} }

type membersDataSource struct {
	client *client.Client
}

type membersDataSourceModel struct {
	Members []memberModel `tfsdk:"members"`
}

type memberModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Email         types.String `tfsdk:"email"`
	GlobalRole    types.String `tfsdk:"global_role"`
	Teams         types.List   `tfsdk:"teams"`
	Environments  types.List   `tfsdk:"environments"`
	LastLoginDate types.String `tfsdk:"last_login_date"`
}

func (d *membersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_members"
}

func (d *membersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Read the list of organization members. Members are managed via the " +
			"GrowthBook invite/SSO flow and are read-only in Terraform.",
		Attributes: map[string]schema.Attribute{
			"members": schema.ListNestedAttribute{
				Computed:    true,
				Description: "All members of the organization.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":              schema.StringAttribute{Computed: true, Description: "Member user ID."},
						"name":            schema.StringAttribute{Computed: true, Description: "Member name."},
						"email":           schema.StringAttribute{Computed: true, Description: "Member email address."},
						"global_role":     schema.StringAttribute{Computed: true, Description: "Organization-wide role."},
						"last_login_date": schema.StringAttribute{Computed: true, Description: "Last login timestamp (RFC3339)."},
						"teams": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Team IDs the member belongs to.",
						},
						"environments": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Environments the member is limited to, if any.",
						},
					},
				},
			},
		},
	}
}

func (d *membersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *membersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	members, err := d.client.ListMembers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list members", err.Error())
		return
	}

	var data membersDataSourceModel
	for _, m := range members {
		data.Members = append(data.Members, memberModel{
			ID:            types.StringValue(m.ID),
			Name:          types.StringValue(m.Name),
			Email:         types.StringValue(m.Email),
			GlobalRole:    types.StringValue(m.GlobalRole),
			Teams:         sliceToStringList(m.Teams),
			Environments:  sliceToStringList(m.Environments),
			LastLoginDate: types.StringValue(m.LastLoginDate),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
