package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*archetypeDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*archetypeDataSource)(nil)
)

func newArchetypeDataSource() datasource.DataSource { return &archetypeDataSource{} }

type archetypeDataSource struct {
	client *client.Client
}

type archetypeDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Owner        types.String `tfsdk:"owner"`
	OwnerEmail   types.String `tfsdk:"owner_email"`
	IsPublic     types.Bool   `tfsdk:"is_public"`
	Attributes   types.String `tfsdk:"attributes"`
	Projects     types.List   `tfsdk:"projects"`
	Environments types.List   `tfsdk:"environments"`
	DateCreated  types.String `tfsdk:"date_created"`
	DateUpdated  types.String `tfsdk:"date_updated"`
}

func (d *archetypeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_archetype"
}

func (d *archetypeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook archetype by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Unique archetype identifier.",
			},
			"name":        schema.StringAttribute{Computed: true, Description: "Archetype name."},
			"description": schema.StringAttribute{Computed: true, Description: "Archetype description."},
			"owner":       schema.StringAttribute{Computed: true, Description: "User ID of the archetype owner."},
			"owner_email": schema.StringAttribute{Computed: true, Description: "Resolved email of the archetype owner."},
			"is_public":   schema.BoolAttribute{Computed: true, Description: "Whether the archetype is shared with the team."},
			"attributes":  schema.StringAttribute{Computed: true, Description: "JSON-encoded object of attribute values."},
			"projects": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Project IDs this archetype is scoped to.",
			},
			"environments": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Environments this archetype is limited to.",
			},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

func (d *archetypeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *archetypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data archetypeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	arch, err := d.client.GetArchetype(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read archetype", err.Error())
		return
	}

	data.Name = types.StringValue(arch.Name)
	data.Description = types.StringValue(arch.Description)
	data.Owner = types.StringValue(arch.Owner)
	data.OwnerEmail = types.StringValue(arch.OwnerEmail)
	data.IsPublic = types.BoolValue(arch.IsPublic)
	data.Attributes = archetypeAttributesToString(arch.Attributes)
	data.Projects = sliceToStringList(arch.Projects)
	data.Environments = sliceToStringList(arch.Environments)
	data.DateCreated = types.StringValue(arch.DateCreated)
	data.DateUpdated = types.StringValue(arch.DateUpdated)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
