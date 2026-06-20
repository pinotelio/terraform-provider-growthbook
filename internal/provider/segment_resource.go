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
	_ resource.Resource                = (*segmentResource)(nil)
	_ resource.ResourceWithConfigure   = (*segmentResource)(nil)
	_ resource.ResourceWithImportState = (*segmentResource)(nil)
)

func newSegmentResource() resource.Resource { return &segmentResource{} }

type segmentResource struct {
	client *client.Client
}

type segmentResourceModel struct {
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

func (r *segmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_segment"
}

func (r *segmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook segment. Segments define a subset of users to filter " +
			"experiment and metric analyses, either via a SQL query (`SQL`) or a fact " +
			"table with filters (`FACT`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique segment identifier (e.g. `seg_...`).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable segment name.",
			},
			"owner": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "User ID or email of the segment owner.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional description of the segment.",
			},
			"datasource_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the data source this segment belongs to.",
			},
			"identifier_type": schema.StringAttribute{
				Required:    true,
				Description: "Identifier type the segment applies to (e.g. `user`, `anonymous`).",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Segment type: `SQL` (query-defined) or `FACT` (fact table + filters).",
			},
			"query": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "SQL query that defines the segment. Required for `SQL` segments.",
			},
			"fact_table_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "ID of the fact table backing the segment. Required for `FACT` segments.",
			},
			"filters": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Fact table filter IDs that further constrain a `FACT` segment.",
			},
			"projects": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Project IDs that can access this segment.",
			},
			"managed_by": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Where the segment is managed from: empty (anywhere) or `api`.",
			},
			"date_created": schema.StringAttribute{
				Computed:      true,
				Description:   "Creation timestamp.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"date_updated": schema.StringAttribute{
				Computed:    true,
				Description: "Last update timestamp.",
			},
		},
	}
}

func (r *segmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (m segmentResourceModel) toInput(ctx context.Context, diags *diagAppender) client.SegmentInput {
	return client.SegmentInput{
		Name:           m.Name.ValueString(),
		Owner:          optString(m.Owner),
		Description:    optString(m.Description),
		DatasourceID:   m.DatasourceID.ValueString(),
		IdentifierType: m.IdentifierType.ValueString(),
		Type:           m.Type.ValueString(),
		Query:          optString(m.Query),
		FactTableID:    optString(m.FactTableID),
		Filters:        diags.strings(ctx, m.Filters),
		Projects:       diags.strings(ctx, m.Projects),
		ManagedBy:      optString(m.ManagedBy),
	}
}

func (r *segmentResource) apply(state *segmentResourceModel, s *client.Segment) {
	state.ID = types.StringValue(s.ID)
	state.Name = types.StringValue(s.Name)
	state.Owner = types.StringValue(s.Owner)
	state.Description = types.StringValue(s.Description)
	state.DatasourceID = types.StringValue(s.DatasourceID)
	state.IdentifierType = types.StringValue(s.IdentifierType)
	state.Type = types.StringValue(s.Type)
	state.Query = types.StringValue(s.Query)
	state.FactTableID = types.StringValue(s.FactTableID)
	state.Filters = sliceToStringList(s.Filters)
	state.Projects = sliceToStringList(s.Projects)
	state.ManagedBy = types.StringValue(s.ManagedBy)
	state.DateCreated = types.StringValue(s.DateCreated)
	state.DateUpdated = types.StringValue(s.DateUpdated)
}

func (r *segmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan segmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateSegment(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create segment", err.Error())
		return
	}
	r.apply(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *segmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state segmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	seg, err := r.client.GetSegment(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read segment", err.Error())
		return
	}
	r.apply(&state, seg)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *segmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan segmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state segmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateSegment(ctx, state.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update segment", err.Error())
		return
	}
	r.apply(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *segmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state segmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteSegment(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete segment", err.Error())
	}
}

func (r *segmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
