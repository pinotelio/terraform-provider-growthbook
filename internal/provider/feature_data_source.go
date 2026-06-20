package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*featureDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*featureDataSource)(nil)
)

func newFeatureDataSource() datasource.DataSource { return &featureDataSource{} }

type featureDataSource struct {
	client *client.Client
}

type featureDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Description  types.String `tfsdk:"description"`
	Owner        types.String `tfsdk:"owner"`
	Project      types.String `tfsdk:"project"`
	Archived     types.Bool   `tfsdk:"archived"`
	ValueType    types.String `tfsdk:"value_type"`
	DefaultValue types.String `tfsdk:"default_value"`
	Tags         types.List   `tfsdk:"tags"`
	DateCreated  types.String `tfsdk:"date_created"`
	DateUpdated  types.String `tfsdk:"date_updated"`
}

func (d *featureDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_feature"
}

func (d *featureDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook feature by key.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Required: true, Description: "Feature key."},
			"description":   schema.StringAttribute{Computed: true, Description: "Feature description."},
			"owner":         schema.StringAttribute{Computed: true, Description: "Feature owner."},
			"project":       schema.StringAttribute{Computed: true, Description: "Associated project ID."},
			"archived":      schema.BoolAttribute{Computed: true, Description: "Whether the feature is archived."},
			"value_type":    schema.StringAttribute{Computed: true, Description: "Feature value type."},
			"default_value": schema.StringAttribute{Computed: true, Description: "Default value."},
			"tags": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags associated with the feature.",
			},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

func (d *featureDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *featureDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data featureDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	feature, err := d.client.GetFeature(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read feature", err.Error())
		return
	}
	data.Description = types.StringValue(feature.Description)
	data.Owner = types.StringValue(feature.Owner)
	data.Project = types.StringValue(feature.Project)
	data.Archived = types.BoolValue(feature.Archived)
	data.ValueType = types.StringValue(feature.ValueType)
	data.DefaultValue = types.StringValue(feature.DefaultValue)
	data.Tags = sliceToStringList(feature.Tags)
	data.DateCreated = types.StringValue(feature.DateCreated)
	data.DateUpdated = types.StringValue(feature.DateUpdated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
