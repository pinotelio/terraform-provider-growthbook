package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*segmentDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*segmentDataSource)(nil)
)

func newSegmentDataSource() datasource.DataSource { return &segmentDataSource{} }

type segmentDataSource struct {
	client *client.Client
}

type segmentDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Owner          types.String `tfsdk:"owner"`
	Description    types.String `tfsdk:"description"`
	DatasourceID   types.String `tfsdk:"datasource_id"`
	IdentifierType types.String `tfsdk:"identifier_type"`
	Type           types.String `tfsdk:"type"`
	Query          types.String `tfsdk:"query"`
	FactTableID    types.String `tfsdk:"fact_table_id"`
	Filters        types.List   `tfsdk:"filters"`
	Projects       types.List   `tfsdk:"projects"`
	ManagedBy      types.String `tfsdk:"managed_by"`
	DateCreated    types.String `tfsdk:"date_created"`
	DateUpdated    types.String `tfsdk:"date_updated"`
}

func (d *segmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_segment"
}

func (d *segmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook segment by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Unique segment identifier (e.g. `seg_...`).",
			},
			"name":            schema.StringAttribute{Computed: true, Description: "Segment name."},
			"owner":           schema.StringAttribute{Computed: true, Description: "Segment owner."},
			"description":     schema.StringAttribute{Computed: true, Description: "Segment description."},
			"datasource_id":   schema.StringAttribute{Computed: true, Description: "Data source ID."},
			"identifier_type": schema.StringAttribute{Computed: true, Description: "Identifier type."},
			"type":            schema.StringAttribute{Computed: true, Description: "Segment type (`SQL` or `FACT`)."},
			"query":           schema.StringAttribute{Computed: true, Description: "SQL query defining the segment."},
			"fact_table_id":   schema.StringAttribute{Computed: true, Description: "Fact table ID for `FACT` segments."},
			"filters": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Fact table filter IDs.",
			},
			"projects": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Project IDs that can access this segment.",
			},
			"managed_by":   schema.StringAttribute{Computed: true, Description: "Where the segment is managed from."},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
		},
	}
}

func (d *segmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *segmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data segmentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	seg, err := d.client.GetSegment(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read segment", err.Error())
		return
	}
	data.Name = types.StringValue(seg.Name)
	data.Owner = types.StringValue(seg.Owner)
	data.Description = types.StringValue(seg.Description)
	data.DatasourceID = types.StringValue(seg.DatasourceID)
	data.IdentifierType = types.StringValue(seg.IdentifierType)
	data.Type = types.StringValue(seg.Type)
	data.Query = types.StringValue(seg.Query)
	data.FactTableID = types.StringValue(seg.FactTableID)
	data.Filters = sliceToStringList(seg.Filters)
	data.Projects = sliceToStringList(seg.Projects)
	data.ManagedBy = types.StringValue(seg.ManagedBy)
	data.DateCreated = types.StringValue(seg.DateCreated)
	data.DateUpdated = types.StringValue(seg.DateUpdated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
