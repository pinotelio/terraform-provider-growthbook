package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*dashboardDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*dashboardDataSource)(nil)
)

func newDashboardDataSource() datasource.DataSource { return &dashboardDataSource{} }

type dashboardDataSource struct {
	client *client.Client
}

func (d *dashboardDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard"
}

func (d *dashboardDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook dashboard by ID.",
		Attributes: map[string]schema.Attribute{
			"id":                  schema.StringAttribute{Required: true, Description: "Unique dashboard identifier."},
			"uid":                 schema.StringAttribute{Computed: true, Description: "Public dashboard UID."},
			"organization":        schema.StringAttribute{Computed: true, Description: "Owning organization ID."},
			"experiment_id":       schema.StringAttribute{Computed: true, Description: "Parent experiment ID."},
			"is_default":          schema.BoolAttribute{Computed: true, Description: "Whether this is the default dashboard."},
			"user_id":             schema.StringAttribute{Computed: true, Description: "Creator user ID."},
			"edit_level":          schema.StringAttribute{Computed: true, Description: "Edit level."},
			"share_level":         schema.StringAttribute{Computed: true, Description: "Share level."},
			"enable_auto_updates": schema.BoolAttribute{Computed: true, Description: "Whether auto-updates are enabled."},
			"update_schedule": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Automatic refresh schedule.",
				Attributes: map[string]schema.Attribute{
					"type":  schema.StringAttribute{Computed: true, Description: "Schedule type."},
					"hours": schema.Float64Attribute{Computed: true, Description: "Staleness threshold in hours."},
					"cron":  schema.StringAttribute{Computed: true, Description: "Cron expression."},
				},
			},
			"title":    schema.StringAttribute{Computed: true, Description: "Display name."},
			"projects": schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Project IDs."},
			"blocks": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Dashboard blocks.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type":          schema.StringAttribute{Computed: true, Description: "Block type."},
						"title":         schema.StringAttribute{Computed: true, Description: "Block title."},
						"description":   schema.StringAttribute{Computed: true, Description: "Block description."},
						"content":       schema.StringAttribute{Computed: true, Description: "Markdown content."},
						"snapshot_id":   schema.StringAttribute{Computed: true, Description: "Associated snapshot ID."},
						"experiment_id": schema.StringAttribute{Computed: true, Description: "Experiment ID."},
						"metric_ids":    schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Metric IDs."},
						"variation_ids": schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Variation IDs."},
						"layout": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Block position on the grid.",
							Attributes: map[string]schema.Attribute{
								"x":      schema.Int64Attribute{Computed: true, Description: "Column position."},
								"y":      schema.Int64Attribute{Computed: true, Description: "Row position."},
								"w":      schema.Int64Attribute{Computed: true, Description: "Width in columns."},
								"h":      schema.Int64Attribute{Computed: true, Description: "Height in rows."},
								"static": schema.BoolAttribute{Computed: true, Description: "Whether the block is fixed."},
							},
						},
					},
				},
			},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

func (d *dashboardDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *dashboardDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dashboardResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	dash, err := d.client.GetDashboard(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read dashboard", err.Error())
		return
	}
	r := &dashboardResource{}
	state := r.apply(dash)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
