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
	_ resource.Resource                = (*savedGroupResource)(nil)
	_ resource.ResourceWithConfigure   = (*savedGroupResource)(nil)
	_ resource.ResourceWithImportState = (*savedGroupResource)(nil)
)

func newSavedGroupResource() resource.Resource { return &savedGroupResource{} }

type savedGroupResource struct {
	client *client.Client
}

type savedGroupResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Type              types.String `tfsdk:"type"`
	Name              types.String `tfsdk:"name"`
	Owner             types.String `tfsdk:"owner"`
	Condition         types.String `tfsdk:"condition"`
	AttributeKey      types.String `tfsdk:"attribute_key"`
	Values            types.List   `tfsdk:"values"`
	Description       types.String `tfsdk:"description"`
	Projects          types.List   `tfsdk:"projects"`
	Archived          types.Bool   `tfsdk:"archived"`
	UseEmptyListGroup types.Bool   `tfsdk:"use_empty_list_group"`
	DateCreated       types.String `tfsdk:"date_created"`
	DateUpdated       types.String `tfsdk:"date_updated"`
}

func (r *savedGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_saved_group"
}

func (r *savedGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook saved group. Saved groups are reusable sets of users, defined " +
			"either as a `list` (an attribute key plus a list of values) or a `condition` " +
			"(a JSON-encoded targeting condition).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique saved group identifier (e.g. `grp_...`).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "The type of saved group: `list` or `condition`. Inferred from the " +
					"other arguments when omitted. Immutable; changing it forces a new saved group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The display name of the saved group.",
			},
			"owner": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The userId or email address of the owner.",
			},
			"condition": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "When `type` is `condition`, the JSON-encoded targeting condition for " +
					"the group.",
			},
			"attribute_key": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "When `type` is `list`, the attribute key the group is based on. " +
					"Immutable; changing it forces a new saved group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"values": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "When `type` is `list`, the list of values for the attribute key.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the saved group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"projects": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Project IDs this saved group is scoped to. Empty means all projects.",
			},
			"archived": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the saved group is archived.",
			},
			"use_empty_list_group": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether an empty list group matches nobody (rather than everybody).",
			},
			"date_created": schema.StringAttribute{
				Computed:      true,
				Description:   "Creation timestamp (RFC3339).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"date_updated": schema.StringAttribute{
				Computed:    true,
				Description: "Last update timestamp (RFC3339).",
			},
		},
	}
}

func (r *savedGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *savedGroupResource) apply(state *savedGroupResourceModel, g *client.SavedGroup) {
	state.ID = types.StringValue(g.ID)
	state.Type = types.StringValue(g.Type)
	state.Name = types.StringValue(g.Name)
	if g.Owner == "" {
		state.Owner = types.StringNull()
	} else {
		state.Owner = types.StringValue(g.Owner)
	}
	if g.Condition == "" {
		state.Condition = types.StringNull()
	} else {
		state.Condition = types.StringValue(g.Condition)
	}
	if g.AttributeKey == "" {
		state.AttributeKey = types.StringNull()
	} else {
		state.AttributeKey = types.StringValue(g.AttributeKey)
	}
	state.Values = sliceToStringList(g.Values)
	state.Description = types.StringValue(g.Description)
	state.Projects = sliceToStringList(g.Projects)
	state.Archived = types.BoolValue(g.Archived)
	state.UseEmptyListGroup = types.BoolValue(g.UseEmptyListGroup)
	state.DateCreated = types.StringValue(g.DateCreated)
	state.DateUpdated = types.StringValue(g.DateUpdated)
}

func (r *savedGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan savedGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.SavedGroupCreateInput{
		Name:         plan.Name.ValueString(),
		Type:         optString(plan.Type),
		Condition:    optString(plan.Condition),
		AttributeKey: optString(plan.AttributeKey),
		Values:       stringListToSlice(ctx, plan.Values, &resp.Diagnostics),
		Owner:        optString(plan.Owner),
		Projects:     stringListToSlice(ctx, plan.Projects, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateSavedGroup(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create saved group", err.Error())
		return
	}
	r.apply(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *savedGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state savedGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.GetSavedGroup(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read saved group", err.Error())
		return
	}
	r.apply(&state, group)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *savedGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan savedGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state savedGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.SavedGroupUpdateInput{
		Name:      optString(plan.Name),
		Condition: optString(plan.Condition),
		Values:    stringListToSlice(ctx, plan.Values, &resp.Diagnostics),
		Owner:     optString(plan.Owner),
		Projects:  stringListToSlice(ctx, plan.Projects, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateSavedGroup(ctx, state.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update saved group", err.Error())
		return
	}
	r.apply(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *savedGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state savedGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteSavedGroup(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete saved group", err.Error())
	}
}

func (r *savedGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
