package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*namespaceDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*namespaceDataSource)(nil)
)

func newNamespaceDataSource() datasource.DataSource { return &namespaceDataSource{} }

type namespaceDataSource struct {
	client *client.Client
}

type namespaceDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	DisplayName   types.String `tfsdk:"display_name"`
	Description   types.String `tfsdk:"description"`
	Status        types.String `tfsdk:"status"`
	Format        types.String `tfsdk:"format"`
	HashAttribute types.String `tfsdk:"hash_attribute"`
	Seed          types.String `tfsdk:"seed"`
}

func (d *namespaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespace"
}

func (d *namespaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook namespace by ID.",
		Attributes: map[string]schema.Attribute{
			"id":             schema.StringAttribute{Required: true, Description: "Unique namespace identifier (e.g. `ns-...`)."},
			"display_name":   schema.StringAttribute{Computed: true, Description: "Human-readable display name."},
			"description":    schema.StringAttribute{Computed: true, Description: "Description of the namespace."},
			"status":         schema.StringAttribute{Computed: true, Description: "Namespace status: `active` or `inactive`."},
			"format":         schema.StringAttribute{Computed: true, Description: "Namespace format: `legacy` or `multiRange`."},
			"hash_attribute": schema.StringAttribute{Computed: true, Description: "The user attribute used for bucket hashing."},
			"seed":           schema.StringAttribute{Computed: true, Description: "The seed used for bucket hashing."},
		},
	}
}

func (d *namespaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *namespaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data namespaceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ns, err := d.client.GetNamespace(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read namespace", err.Error())
		return
	}

	data.DisplayName = types.StringValue(ns.DisplayName)
	data.Description = types.StringValue(ns.Description)
	data.Status = types.StringValue(ns.Status)
	data.Format = types.StringValue(ns.Format)
	if ns.HashAttribute == "" {
		data.HashAttribute = types.StringNull()
	} else {
		data.HashAttribute = types.StringValue(ns.HashAttribute)
	}
	if ns.Seed == "" {
		data.Seed = types.StringNull()
	} else {
		data.Seed = types.StringValue(ns.Seed)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
