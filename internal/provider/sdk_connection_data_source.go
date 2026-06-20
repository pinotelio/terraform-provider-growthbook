package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*sdkConnectionDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*sdkConnectionDataSource)(nil)
)

func newSDKConnectionDataSource() datasource.DataSource { return &sdkConnectionDataSource{} }

type sdkConnectionDataSource struct {
	client *client.Client
}

type sdkConnectionDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Languages   types.List   `tfsdk:"languages"`
	Environment types.String `tfsdk:"environment"`
	Projects    types.List   `tfsdk:"projects"`
	Key         types.String `tfsdk:"key"`
	SSEEnabled  types.Bool   `tfsdk:"sse_enabled"`
	ProxyHost   types.String `tfsdk:"proxy_host"`
}

func (d *sdkConnectionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sdk_connection"
}

func (d *sdkConnectionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook SDK connection by ID.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Required: true, Description: "SDK connection ID."},
			"name":        schema.StringAttribute{Computed: true, Description: "Connection name."},
			"environment": schema.StringAttribute{Computed: true, Description: "Environment served."},
			"key":         schema.StringAttribute{Computed: true, Description: "Client key used by SDKs."},
			"sse_enabled": schema.BoolAttribute{Computed: true, Description: "Whether SSE streaming is enabled."},
			"proxy_host":  schema.StringAttribute{Computed: true, Description: "Proxy host URL, if any."},
			"languages": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Languages associated with the connection.",
			},
			"projects": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Project IDs the connection is scoped to.",
			},
		},
	}
}

func (d *sdkConnectionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *sdkConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data sdkConnectionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	conn, err := d.client.GetSDKConnection(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read SDK connection", err.Error())
		return
	}
	data.Name = types.StringValue(conn.Name)
	data.Languages = sliceToStringList(conn.Languages)
	data.Environment = types.StringValue(conn.Environment)
	data.Projects = sliceToStringList(conn.Projects)
	data.Key = types.StringValue(conn.Key)
	data.SSEEnabled = types.BoolValue(conn.SSEEnabled)
	data.ProxyHost = types.StringValue(conn.ProxyHost)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
