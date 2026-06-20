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
	_ resource.Resource                = (*dashboardResource)(nil)
	_ resource.ResourceWithConfigure   = (*dashboardResource)(nil)
	_ resource.ResourceWithImportState = (*dashboardResource)(nil)
)

func newDashboardResource() resource.Resource { return &dashboardResource{} }

type dashboardResource struct {
	client *client.Client
}

type dashboardUpdateScheduleModel struct {
	Type  types.String  `tfsdk:"type"`
	Hours types.Float64 `tfsdk:"hours"`
	Cron  types.String  `tfsdk:"cron"`
}

type dashboardBlockLayoutModel struct {
	X      types.Int64 `tfsdk:"x"`
	Y      types.Int64 `tfsdk:"y"`
	W      types.Int64 `tfsdk:"w"`
	H      types.Int64 `tfsdk:"h"`
	Static types.Bool  `tfsdk:"static"`
}

type dashboardBlockModel struct {
	Type         types.String               `tfsdk:"type"`
	Title        types.String               `tfsdk:"title"`
	Description  types.String               `tfsdk:"description"`
	Content      types.String               `tfsdk:"content"`
	SnapshotID   types.String               `tfsdk:"snapshot_id"`
	ExperimentID types.String               `tfsdk:"experiment_id"`
	MetricIDs    types.List                 `tfsdk:"metric_ids"`
	VariationIDs types.List                 `tfsdk:"variation_ids"`
	Layout       *dashboardBlockLayoutModel `tfsdk:"layout"`
}

type dashboardResourceModel struct {
	ID                types.String                  `tfsdk:"id"`
	UID               types.String                  `tfsdk:"uid"`
	Organization      types.String                  `tfsdk:"organization"`
	ExperimentID      types.String                  `tfsdk:"experiment_id"`
	IsDefault         types.Bool                    `tfsdk:"is_default"`
	UserID            types.String                  `tfsdk:"user_id"`
	EditLevel         types.String                  `tfsdk:"edit_level"`
	ShareLevel        types.String                  `tfsdk:"share_level"`
	EnableAutoUpdates types.Bool                    `tfsdk:"enable_auto_updates"`
	UpdateSchedule    *dashboardUpdateScheduleModel `tfsdk:"update_schedule"`
	Title             types.String                  `tfsdk:"title"`
	Projects          types.List                    `tfsdk:"projects"`
	Blocks            []dashboardBlockModel         `tfsdk:"blocks"`
	DateCreated       types.String                  `tfsdk:"date_created"`
	DateUpdated       types.String                  `tfsdk:"date_updated"`
}

func (r *dashboardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard"
}

func (r *dashboardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook dashboard. Dashboards are composed of blocks and can be tied to " +
			"an experiment or stand alone as a general dashboard.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique dashboard identifier.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"uid":          schema.StringAttribute{Computed: true, Description: "Public dashboard UID."},
			"organization": schema.StringAttribute{Computed: true, Description: "Owning organization ID."},
			"experiment_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Parent experiment ID for an experiment dashboard. Omit for a general dashboard. Immutable; changing it forces a new dashboard.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_default": schema.BoolAttribute{Computed: true, Description: "Whether this is the default dashboard."},
			"user_id":    schema.StringAttribute{Computed: true, Description: "ID of the user who created the dashboard."},
			"edit_level": schema.StringAttribute{
				Required:    true,
				Description: "Edit level: `published` or `private`.",
			},
			"share_level": schema.StringAttribute{
				Required:    true,
				Description: "Share level: `published` or `private`.",
			},
			"enable_auto_updates": schema.BoolAttribute{
				Required:    true,
				Description: "Whether automatic updates are enabled (general dashboards require an update_schedule).",
			},
			"update_schedule": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Automatic refresh schedule for general dashboards. Write-only: not refreshed from the server.",
				Attributes: map[string]schema.Attribute{
					"type":  schema.StringAttribute{Required: true, Description: "Schedule type: `stale` or `cron`."},
					"hours": schema.Float64Attribute{Optional: true, Description: "Staleness threshold in hours (for `stale`)."},
					"cron":  schema.StringAttribute{Optional: true, Description: "Cron expression (for `cron`)."},
				},
			},
			"title": schema.StringAttribute{
				Required:    true,
				Description: "Display name of the dashboard.",
			},
			"projects": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Project IDs (general dashboards only).",
			},
			"blocks": schema.ListNestedAttribute{
				Optional: true,
				Description: "Ordered list of dashboard blocks. Currently only `markdown` blocks are fully " +
					"supported by this provider; other block types require type-specific fields that the API " +
					"rejects when absent. Leave empty for a blank dashboard.",
				NestedObject: dashboardBlockSchema(),
			},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

func dashboardBlockSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Required: true,
				Description: "Block type. Use `markdown` (the only block type with full provider support); " +
					"other GrowthBook block types require type-specific fields not modeled here.",
			},
			"title":       schema.StringAttribute{Required: true, Description: "Block title."},
			"description": schema.StringAttribute{Optional: true, Description: "Block description."},
			"content":     schema.StringAttribute{Optional: true, Description: "Markdown content (required for `markdown` blocks)."},
			"snapshot_id": schema.StringAttribute{Optional: true, Description: "Associated snapshot ID."},
			"experiment_id": schema.StringAttribute{
				Optional:    true,
				Description: "Experiment ID (experiment-related blocks).",
			},
			"metric_ids": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Metric IDs (metric blocks).",
			},
			"variation_ids": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Variation IDs (experiment blocks).",
			},
			"layout": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Position of the block on the grid.",
				Attributes: map[string]schema.Attribute{
					"x":      schema.Int64Attribute{Required: true, Description: "Column position (0-23)."},
					"y":      schema.Int64Attribute{Required: true, Description: "Row position."},
					"w":      schema.Int64Attribute{Required: true, Description: "Width in columns (1-24)."},
					"h":      schema.Int64Attribute{Required: true, Description: "Height in rows."},
					"static": schema.BoolAttribute{Optional: true, Description: "Whether the block is fixed in place."},
				},
			},
		},
	}
}

func (r *dashboardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (m dashboardResourceModel) toInput(ctx context.Context, da *diagAppender) client.DashboardInput {
	in := client.DashboardInput{
		Title:             m.Title.ValueString(),
		EditLevel:         m.EditLevel.ValueString(),
		ShareLevel:        m.ShareLevel.ValueString(),
		EnableAutoUpdates: m.EnableAutoUpdates.ValueBool(),
		ExperimentID:      optString(m.ExperimentID),
		Projects:          da.strings(ctx, m.Projects),
		Blocks:            make([]client.DashboardBlock, 0, len(m.Blocks)),
	}
	if m.UpdateSchedule != nil {
		in.UpdateSchedule = &client.DashboardUpdateSchedule{
			Type:  m.UpdateSchedule.Type.ValueString(),
			Hours: optFloat64(m.UpdateSchedule.Hours),
			Cron:  m.UpdateSchedule.Cron.ValueString(),
		}
	}
	for _, b := range m.Blocks {
		block := client.DashboardBlock{
			Type:         b.Type.ValueString(),
			Title:        b.Title.ValueString(),
			Description:  b.Description.ValueString(),
			Content:      optString(b.Content),
			SnapshotID:   optString(b.SnapshotID),
			ExperimentID: optString(b.ExperimentID),
			MetricIDs:    da.strings(ctx, b.MetricIDs),
			VariationIDs: da.strings(ctx, b.VariationIDs),
		}
		if b.Layout != nil {
			block.Layout = &client.DashboardBlockLayout{
				X:      b.Layout.X.ValueInt64(),
				Y:      b.Layout.Y.ValueInt64(),
				W:      b.Layout.W.ValueInt64(),
				H:      b.Layout.H.ValueInt64(),
				Static: optBool(b.Layout.Static),
			}
		}
		in.Blocks = append(in.Blocks, block)
	}
	return in
}

func (r *dashboardResource) apply(d *client.Dashboard) dashboardResourceModel {
	state := dashboardResourceModel{
		ID:                types.StringValue(d.ID),
		UID:               types.StringValue(d.UID),
		Organization:      types.StringValue(d.Organization),
		ExperimentID:      types.StringValue(d.ExperimentID),
		IsDefault:         types.BoolValue(d.IsDefault),
		UserID:            types.StringValue(d.UserID),
		EditLevel:         types.StringValue(d.EditLevel),
		ShareLevel:        types.StringValue(d.ShareLevel),
		EnableAutoUpdates: types.BoolValue(d.EnableAutoUpdates),
		Title:             types.StringValue(d.Title),
		Projects:          sliceToStringList(d.Projects),
		DateCreated:       types.StringValue(d.DateCreated),
		DateUpdated:       types.StringValue(d.DateUpdated),
	}
	// update_schedule and blocks are write-only: they are preserved from
	// configuration (Create/Update) or prior state (Read) rather than read back,
	// because the server normalizes/auto-populates block layout and schedule
	// fields that would otherwise produce inconsistent-result errors.
	return state
}

func (r *dashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateDashboard(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create dashboard", err.Error())
		return
	}
	state := r.apply(created)
	// Preserve the configured write-only attributes verbatim (see apply): the API
	// may normalize/auto-populate blocks and the schedule, which would otherwise
	// raise an inconsistent-result error.
	state.Blocks = plan.Blocks
	state.UpdateSchedule = plan.UpdateSchedule
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dashboardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	d, err := r.client.GetDashboard(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read dashboard", err.Error())
		return
	}
	newState := r.apply(d)
	// Preserve write-only attributes from prior state (not read back).
	newState.Blocks = state.Blocks
	newState.UpdateSchedule = state.UpdateSchedule
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *dashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state dashboardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateDashboard(ctx, state.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update dashboard", err.Error())
		return
	}
	newState := r.apply(updated)
	// Preserve configured write-only attributes verbatim (see Create).
	newState.Blocks = plan.Blocks
	newState.UpdateSchedule = plan.UpdateSchedule
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *dashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dashboardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteDashboard(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete dashboard", err.Error())
	}
}

func (r *dashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
