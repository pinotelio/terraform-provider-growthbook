package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*customFieldDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*customFieldDataSource)(nil)
)

func newCustomFieldDataSource() datasource.DataSource { return &customFieldDataSource{} }

type customFieldDataSource struct {
	client *client.Client
}

type customFieldDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Placeholder  types.String `tfsdk:"placeholder"`
	DefaultValue types.String `tfsdk:"default_value"`
	Type         types.String `tfsdk:"type"`
	Values       types.String `tfsdk:"values"`
	Required     types.Bool   `tfsdk:"required"`
	Creator      types.String `tfsdk:"creator"`
	Projects     types.List   `tfsdk:"projects"`
	Sections     types.List   `tfsdk:"sections"`
	Active       types.Bool   `tfsdk:"active"`
	DateCreated  types.String `tfsdk:"date_created"`
	DateUpdated  types.String `tfsdk:"date_updated"`
}

func (d *customFieldDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_field"
}

func (d *customFieldDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook custom field by ID (its unique key).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Unique key of the custom field.",
			},
			"name":          schema.StringAttribute{Computed: true, Description: "Display name."},
			"description":   schema.StringAttribute{Computed: true, Description: "Description."},
			"placeholder":   schema.StringAttribute{Computed: true, Description: "Placeholder text."},
			"default_value": schema.StringAttribute{Computed: true, Description: "JSON-encoded default value."},
			"type":          schema.StringAttribute{Computed: true, Description: "Value type."},
			"values":        schema.StringAttribute{Computed: true, Description: "Comma-separated allowed values."},
			"required":      schema.BoolAttribute{Computed: true, Description: "Whether a value is required."},
			"creator":       schema.StringAttribute{Computed: true, Description: "User ID of the creator."},
			"projects": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Project IDs this field applies to.",
			},
			"sections": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Object types this field applies to.",
			},
			"active":       schema.BoolAttribute{Computed: true, Description: "Whether the field is active."},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

func (d *customFieldDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *customFieldDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data customFieldDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	field, err := d.client.GetCustomField(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read custom field", err.Error())
		return
	}

	data.Name = types.StringValue(field.Name)
	data.Description = types.StringValue(field.Description)
	data.Placeholder = types.StringValue(field.Placeholder)
	data.DefaultValue = customFieldDefaultValueToString(field.DefaultValue)
	data.Type = types.StringValue(field.Type)
	data.Values = types.StringValue(field.Values)
	data.Required = types.BoolValue(field.Required)
	data.Creator = types.StringValue(field.Creator)
	data.Projects = sliceToStringList(field.Projects)
	data.Sections = sliceToStringList(field.Sections)
	data.Active = types.BoolValue(field.Active)
	data.DateCreated = types.StringValue(field.DateCreated)
	data.DateUpdated = types.StringValue(field.DateUpdated)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
