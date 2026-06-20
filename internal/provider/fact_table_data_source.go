package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*factTableDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*factTableDataSource)(nil)
)

func newFactTableDataSource() datasource.DataSource { return &factTableDataSource{} }

type factTableDataSource struct {
	client *client.Client
}

func (d *factTableDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fact_table"
}

func (d *factTableDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook fact table by ID.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Required: true, Description: "Unique fact table identifier."},
			"name":        schema.StringAttribute{Computed: true, Description: "Fact table name."},
			"description": schema.StringAttribute{Computed: true, Description: "Fact table description."},
			"owner":       schema.StringAttribute{Computed: true, Description: "Owner userId or email address."},
			"projects": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Associated project IDs.",
			},
			"tags": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags associated with the fact table.",
			},
			"datasource": schema.StringAttribute{Computed: true, Description: "Data source ID."},
			"user_id_types": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Identifier columns available in the query.",
			},
			"sql":        schema.StringAttribute{Computed: true, Description: "SQL query that defines the fact table."},
			"event_name": schema.StringAttribute{Computed: true, Description: "Event name used in SQL template variables."},
			"columns": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Column definitions derived from the SQL.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"column":               schema.StringAttribute{Computed: true, Description: "Actual column name in the SQL query."},
						"datatype":             schema.StringAttribute{Computed: true, Description: "Column data type."},
						"number_format":        schema.StringAttribute{Computed: true, Description: "Number format hint."},
						"name":                 schema.StringAttribute{Computed: true, Description: "Display name for the column."},
						"description":          schema.StringAttribute{Computed: true, Description: "Column description."},
						"always_inline_filter": schema.BoolAttribute{Computed: true, Description: "Always include this column as an inline filter."},
						"deleted":              schema.BoolAttribute{Computed: true, Description: "Whether the column has been deleted."},
						"is_auto_slice_column": schema.BoolAttribute{Computed: true, Description: "Whether this column supports auto slice analysis."},
						"auto_slices": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Specific slices to automatically analyze.",
						},
						"locked_auto_slices": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Slices protected from automatic updates.",
						},
						"date_created": schema.StringAttribute{Computed: true, Description: "Column creation timestamp (RFC3339)."},
						"date_updated": schema.StringAttribute{Computed: true, Description: "Column last update timestamp (RFC3339)."},
					},
				},
			},
			"archived":     schema.BoolAttribute{Computed: true, Description: "Whether the fact table is archived."},
			"managed_by":   schema.StringAttribute{Computed: true, Description: "Where the fact table is managed from."},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

func (d *factTableDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *factTableDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data factTableResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ft, err := d.client.GetFactTable(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read fact table", err.Error())
		return
	}
	(&factTableResource{}).apply(&data, ft)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
