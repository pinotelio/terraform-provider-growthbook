package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*dimensionDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*dimensionDataSource)(nil)
)

func newDimensionDataSource() datasource.DataSource { return &dimensionDataSource{} }

type dimensionDataSource struct {
	client *client.Client
}

type dimensionDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Owner          types.String `tfsdk:"owner"`
	Description    types.String `tfsdk:"description"`
	DatasourceID   types.String `tfsdk:"datasource_id"`
	IdentifierType types.String `tfsdk:"identifier_type"`
	Query          types.String `tfsdk:"query"`
	ManagedBy      types.String `tfsdk:"managed_by"`
	DateCreated    types.String `tfsdk:"date_created"`
	DateUpdated    types.String `tfsdk:"date_updated"`
}

func (d *dimensionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dimension"
}

func (d *dimensionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook dimension by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Unique dimension identifier (e.g. `dim_...`).",
			},
			"name":            schema.StringAttribute{Computed: true, Description: "Dimension name."},
			"owner":           schema.StringAttribute{Computed: true, Description: "Dimension owner."},
			"description":     schema.StringAttribute{Computed: true, Description: "Dimension description."},
			"datasource_id":   schema.StringAttribute{Computed: true, Description: "Data source ID."},
			"identifier_type": schema.StringAttribute{Computed: true, Description: "Identifier type."},
			"query":           schema.StringAttribute{Computed: true, Description: "SQL query defining the dimension."},
			"managed_by":      schema.StringAttribute{Computed: true, Description: "Where the dimension is managed from."},
			"date_created":    schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
			"date_updated":    schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
		},
	}
}

func (d *dimensionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *dimensionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dimensionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	dim, err := d.client.GetDimension(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read dimension", err.Error())
		return
	}
	data.Name = types.StringValue(dim.Name)
	data.Owner = types.StringValue(dim.Owner)
	data.Description = types.StringValue(dim.Description)
	data.DatasourceID = types.StringValue(dim.DatasourceID)
	data.IdentifierType = types.StringValue(dim.IdentifierType)
	data.Query = types.StringValue(dim.Query)
	data.ManagedBy = types.StringValue(dim.ManagedBy)
	data.DateCreated = types.StringValue(dim.DateCreated)
	data.DateUpdated = types.StringValue(dim.DateUpdated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
