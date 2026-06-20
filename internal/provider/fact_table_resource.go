package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ resource.Resource                = (*factTableResource)(nil)
	_ resource.ResourceWithConfigure   = (*factTableResource)(nil)
	_ resource.ResourceWithImportState = (*factTableResource)(nil)
)

func newFactTableResource() resource.Resource { return &factTableResource{} }

type factTableResource struct {
	client *client.Client
}

type factTableResourceModel struct {
	ID          types.String           `tfsdk:"id"`
	Name        types.String           `tfsdk:"name"`
	Description types.String           `tfsdk:"description"`
	Owner       types.String           `tfsdk:"owner"`
	Projects    types.List             `tfsdk:"projects"`
	Tags        types.List             `tfsdk:"tags"`
	Datasource  types.String           `tfsdk:"datasource"`
	UserIDTypes types.List             `tfsdk:"user_id_types"`
	SQL         types.String           `tfsdk:"sql"`
	EventName   types.String           `tfsdk:"event_name"`
	Columns     []factTableColumnModel `tfsdk:"columns"`
	Archived    types.Bool             `tfsdk:"archived"`
	ManagedBy   types.String           `tfsdk:"managed_by"`
	DateCreated types.String           `tfsdk:"date_created"`
	DateUpdated types.String           `tfsdk:"date_updated"`
}

type factTableColumnModel struct {
	Column             types.String `tfsdk:"column"`
	Datatype           types.String `tfsdk:"datatype"`
	NumberFormat       types.String `tfsdk:"number_format"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	AlwaysInlineFilter types.Bool   `tfsdk:"always_inline_filter"`
	Deleted            types.Bool   `tfsdk:"deleted"`
	IsAutoSliceColumn  types.Bool   `tfsdk:"is_auto_slice_column"`
	AutoSlices         types.List   `tfsdk:"auto_slices"`
	LockedAutoSlices   types.List   `tfsdk:"locked_auto_slices"`
}

func (r *factTableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fact_table"
}

func (r *factTableResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *factTableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *factTableResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook fact table. Fact tables define a SQL query against a " +
			"data source that fact metrics and filters are built on top of.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique fact table identifier.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable fact table name.",
			},
			"description": schema.StringAttribute{Optional: true, Computed: true, Description: "Description of the fact table."},
			"owner":       schema.StringAttribute{Optional: true, Computed: true, Description: "Owner userId or email address."},
			"projects": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Associated project IDs.",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags associated with the fact table.",
			},
			"datasource": schema.StringAttribute{
				Required:      true,
				Description:   "Data source ID this fact table queries. Changing this forces a new fact table.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"user_id_types": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Identifier columns available in the query (e.g. `id`, `anonymous_id`).",
			},
			"sql": schema.StringAttribute{
				Required:    true,
				Description: "SQL query that defines the fact table.",
			},
			"event_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Event name used in SQL template variables.",
			},
			"columns": schema.ListNestedAttribute{
				Optional: true,
				Description: "Column metadata overrides. Columns are derived by GrowthBook from parsing the SQL; " +
					"use this to set metadata (datatype, display name, etc.) on derived columns. Write-only: " +
					"the column set is not refreshed from the server, so only the columns you declare are tracked.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"column":   schema.StringAttribute{Required: true, Description: "Actual column name in the SQL query."},
						"datatype": schema.StringAttribute{Optional: true, Description: "Column data type: `number`, `string`, `date`, `boolean`, `json`, `binary`, `other`, or empty."},
						"number_format": schema.StringAttribute{
							Optional:    true,
							Description: "Number format hint: empty, `currency`, `time:seconds`, `memory:bytes`, or `memory:kilobytes`.",
						},
						"name":                 schema.StringAttribute{Optional: true, Description: "Display name for the column."},
						"description":          schema.StringAttribute{Optional: true, Description: "Column description."},
						"always_inline_filter": schema.BoolAttribute{Optional: true, Description: "Always include this column as an inline filter in queries."},
						"deleted":              schema.BoolAttribute{Optional: true, Description: "Whether the column has been deleted from the source."},
						"is_auto_slice_column": schema.BoolAttribute{Optional: true, Description: "Whether this column can be used for auto slice analysis (enterprise)."},
						"auto_slices": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Specific slices to automatically analyze for this column.",
						},
						"locked_auto_slices": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Slices protected from automatic updates.",
						},
					},
				},
			},
			"archived": schema.BoolAttribute{Optional: true, Computed: true, Description: "Whether the fact table is archived."},
			"managed_by": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Where the fact table is managed from: empty, `api`, or `admin`. Set to `api` to disable editing in the GrowthBook UI.",
			},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

func factTableColumnsToClient(ctx context.Context, columns []factTableColumnModel, diags *diagAppender) []client.FactTableColumn {
	if len(columns) == 0 {
		return nil
	}
	out := make([]client.FactTableColumn, 0, len(columns))
	for _, col := range columns {
		out = append(out, client.FactTableColumn{
			Column:             col.Column.ValueString(),
			Datatype:           col.Datatype.ValueString(),
			NumberFormat:       col.NumberFormat.ValueString(),
			Name:               col.Name.ValueString(),
			Description:        col.Description.ValueString(),
			AlwaysInlineFilter: col.AlwaysInlineFilter.ValueBool(),
			Deleted:            col.Deleted.ValueBool(),
			IsAutoSliceColumn:  col.IsAutoSliceColumn.ValueBool(),
			AutoSlices:         diags.strings(ctx, col.AutoSlices),
			LockedAutoSlices:   diags.strings(ctx, col.LockedAutoSlices),
		})
	}
	return out
}

func (m factTableResourceModel) toCreateInput(ctx context.Context, diags *diagAppender) client.FactTableCreateInput {
	return client.FactTableCreateInput{
		Name:        m.Name.ValueString(),
		Description: optString(m.Description),
		Owner:       optString(m.Owner),
		Projects:    diags.strings(ctx, m.Projects),
		Tags:        diags.strings(ctx, m.Tags),
		Datasource:  m.Datasource.ValueString(),
		UserIDTypes: diags.strings(ctx, m.UserIDTypes),
		SQL:         m.SQL.ValueString(),
		EventName:   optString(m.EventName),
		ManagedBy:   optString(m.ManagedBy),
	}
}

func (m factTableResourceModel) toUpdateInput(ctx context.Context, diags *diagAppender) client.FactTableUpdateInput {
	return client.FactTableUpdateInput{
		Name:        m.Name.ValueString(),
		Description: optString(m.Description),
		Owner:       optString(m.Owner),
		Projects:    diags.strings(ctx, m.Projects),
		Tags:        diags.strings(ctx, m.Tags),
		UserIDTypes: diags.strings(ctx, m.UserIDTypes),
		SQL:         m.SQL.ValueString(),
		EventName:   optString(m.EventName),
		Columns:     factTableColumnsToClient(ctx, m.Columns, diags),
		ManagedBy:   optString(m.ManagedBy),
		Archived:    optBool(m.Archived),
	}
}

func (r *factTableResource) apply(state *factTableResourceModel, ft *client.FactTable) {
	state.ID = types.StringValue(ft.ID)
	state.Name = types.StringValue(ft.Name)
	state.Description = types.StringValue(ft.Description)
	state.Owner = types.StringValue(ft.Owner)
	state.Projects = sliceToStringList(ft.Projects)
	state.Tags = sliceToStringList(ft.Tags)
	state.Datasource = types.StringValue(ft.Datasource)
	state.UserIDTypes = sliceToStringList(ft.UserIDTypes)
	state.SQL = types.StringValue(ft.SQL)
	state.EventName = types.StringValue(ft.EventName)
	// columns are write-only (preserved from configuration, not read back): the
	// server derives the full column set from the SQL, which would otherwise
	// conflict with the subset of columns the practitioner declares metadata for.
	state.Archived = types.BoolValue(ft.Archived)
	state.ManagedBy = types.StringValue(ft.ManagedBy)
	state.DateCreated = types.StringValue(ft.DateCreated)
	state.DateUpdated = types.StringValue(ft.DateUpdated)
}

func (r *factTableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan factTableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toCreateInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateFactTable(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create fact table", err.Error())
		return
	}
	r.apply(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *factTableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state factTableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ft, err := r.client.GetFactTable(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read fact table", err.Error())
		return
	}
	r.apply(&state, ft)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *factTableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan factTableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state factTableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toUpdateInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateFactTable(ctx, state.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update fact table", err.Error())
		return
	}
	r.apply(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *factTableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state factTableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteFactTable(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete fact table", err.Error())
	}
}
