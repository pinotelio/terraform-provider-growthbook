package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*attributeDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*attributeDataSource)(nil)
)

func newAttributeDataSource() datasource.DataSource { return &attributeDataSource{} }

type attributeDataSource struct {
	client *client.Client
}

type attributeDataSourceModel struct {
	Property      types.String `tfsdk:"property"`
	Datatype      types.String `tfsdk:"datatype"`
	Description   types.String `tfsdk:"description"`
	HashAttribute types.Bool   `tfsdk:"hash_attribute"`
	Archived      types.Bool   `tfsdk:"archived"`
	Enum          types.String `tfsdk:"enum"`
	Format        types.String `tfsdk:"format"`
	Projects      types.List   `tfsdk:"projects"`
	Tags          types.List   `tfsdk:"tags"`
}

func (d *attributeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_attribute"
}

func (d *attributeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook targeting attribute by property.",
		Attributes: map[string]schema.Attribute{
			"property":       schema.StringAttribute{Required: true, Description: "The attribute property name (its unique identifier)."},
			"datatype":       schema.StringAttribute{Computed: true, Description: "The attribute datatype."},
			"description":    schema.StringAttribute{Computed: true, Description: "Description of the attribute."},
			"hash_attribute": schema.BoolAttribute{Computed: true, Description: "Whether the attribute is hashed for bucketing."},
			"archived":       schema.BoolAttribute{Computed: true, Description: "Whether the attribute is archived."},
			"enum":           schema.StringAttribute{Computed: true, Description: "Comma-separated allowed values when datatype is `enum`."},
			"format":         schema.StringAttribute{Computed: true, Description: "The attribute's format."},
			"projects": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Project IDs this attribute is scoped to.",
			},
			"tags": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags applied to the attribute.",
			},
		},
	}
}

func (d *attributeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *attributeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data attributeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attr, err := d.client.GetAttribute(ctx, data.Property.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read attribute", err.Error())
		return
	}

	data.Datatype = types.StringValue(attr.Datatype)
	data.Description = types.StringValue(attr.Description)
	data.HashAttribute = types.BoolValue(attr.HashAttribute)
	data.Archived = types.BoolValue(attr.Archived)
	data.Enum = types.StringValue(attr.Enum)
	data.Format = types.StringValue(attr.Format)
	data.Projects = sliceToStringList(attr.Projects)
	data.Tags = sliceToStringList(attr.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
