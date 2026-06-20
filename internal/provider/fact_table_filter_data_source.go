package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*factTableFilterDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*factTableFilterDataSource)(nil)
)

func newFactTableFilterDataSource() datasource.DataSource { return &factTableFilterDataSource{} }

type factTableFilterDataSource struct {
	client *client.Client
}

func (d *factTableFilterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fact_table_filter"
}

func (d *factTableFilterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook fact table filter by fact table ID and filter ID.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Required: true, Description: "Unique fact table filter identifier."},
			"fact_table_id": schema.StringAttribute{Required: true, Description: "ID of the fact table the filter belongs to."},
			"name":          schema.StringAttribute{Computed: true, Description: "Filter name."},
			"description":   schema.StringAttribute{Computed: true, Description: "Filter description."},
			"value":         schema.StringAttribute{Computed: true, Description: "SQL expression for this filter."},
			"managed_by":    schema.StringAttribute{Computed: true, Description: "Where the filter is managed from."},
			"date_created":  schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated":  schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

func (d *factTableFilterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *factTableFilterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data factTableFilterResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	factTableID := data.FactTableID.ValueString()
	f, err := d.client.GetFactTableFilter(ctx, factTableID, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read fact table filter", err.Error())
		return
	}
	(&factTableFilterResource{}).apply(&data, factTableID, f)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
