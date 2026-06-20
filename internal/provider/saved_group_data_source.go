package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*savedGroupDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*savedGroupDataSource)(nil)
)

func newSavedGroupDataSource() datasource.DataSource { return &savedGroupDataSource{} }

type savedGroupDataSource struct {
	client *client.Client
}

type savedGroupDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Type              types.String `tfsdk:"type"`
	Name              types.String `tfsdk:"name"`
	Owner             types.String `tfsdk:"owner"`
	Condition         types.String `tfsdk:"condition"`
	AttributeKey      types.String `tfsdk:"attribute_key"`
	Values            types.List   `tfsdk:"values"`
	Description       types.String `tfsdk:"description"`
	Projects          types.List   `tfsdk:"projects"`
	Archived          types.Bool   `tfsdk:"archived"`
	UseEmptyListGroup types.Bool   `tfsdk:"use_empty_list_group"`
	DateCreated       types.String `tfsdk:"date_created"`
	DateUpdated       types.String `tfsdk:"date_updated"`
}

func (d *savedGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_saved_group"
}

func (d *savedGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook saved group by ID.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Required: true, Description: "Unique saved group identifier (e.g. `grp_...`)."},
			"type":          schema.StringAttribute{Computed: true, Description: "The type of saved group: `list` or `condition`."},
			"name":          schema.StringAttribute{Computed: true, Description: "The display name of the saved group."},
			"owner":         schema.StringAttribute{Computed: true, Description: "The userId or email address of the owner."},
			"condition":     schema.StringAttribute{Computed: true, Description: "JSON-encoded targeting condition (condition groups)."},
			"attribute_key": schema.StringAttribute{Computed: true, Description: "The attribute key the group is based on (list groups)."},
			"values": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The list of values for the attribute key (list groups).",
			},
			"description": schema.StringAttribute{Computed: true, Description: "Description of the saved group."},
			"projects": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Project IDs this saved group is scoped to.",
			},
			"archived":             schema.BoolAttribute{Computed: true, Description: "Whether the saved group is archived."},
			"use_empty_list_group": schema.BoolAttribute{Computed: true, Description: "Whether an empty list group matches nobody."},
			"date_created":         schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated":         schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

func (d *savedGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *savedGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data savedGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := d.client.GetSavedGroup(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read saved group", err.Error())
		return
	}

	data.Type = types.StringValue(group.Type)
	data.Name = types.StringValue(group.Name)
	if group.Owner == "" {
		data.Owner = types.StringNull()
	} else {
		data.Owner = types.StringValue(group.Owner)
	}
	if group.Condition == "" {
		data.Condition = types.StringNull()
	} else {
		data.Condition = types.StringValue(group.Condition)
	}
	if group.AttributeKey == "" {
		data.AttributeKey = types.StringNull()
	} else {
		data.AttributeKey = types.StringValue(group.AttributeKey)
	}
	data.Values = sliceToStringList(group.Values)
	data.Description = types.StringValue(group.Description)
	data.Projects = sliceToStringList(group.Projects)
	data.Archived = types.BoolValue(group.Archived)
	data.UseEmptyListGroup = types.BoolValue(group.UseEmptyListGroup)
	data.DateCreated = types.StringValue(group.DateCreated)
	data.DateUpdated = types.StringValue(group.DateUpdated)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
