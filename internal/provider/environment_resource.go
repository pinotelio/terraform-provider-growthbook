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
	_ resource.Resource                = (*environmentResource)(nil)
	_ resource.ResourceWithConfigure   = (*environmentResource)(nil)
	_ resource.ResourceWithImportState = (*environmentResource)(nil)
)

func newEnvironmentResource() resource.Resource { return &environmentResource{} }

type environmentResource struct {
	client *client.Client
}

type environmentResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Description  types.String `tfsdk:"description"`
	ToggleOnList types.Bool   `tfsdk:"toggle_on_list"`
	DefaultState types.Bool   `tfsdk:"default_state"`
	Projects     types.List   `tfsdk:"projects"`
	Parent       types.String `tfsdk:"parent"`
}

func (r *environmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *environmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook environment (e.g. `production`, `staging`). Environments " +
			"control where feature flag rules are evaluated.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The environment identifier/name. Changing this forces a new environment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the environment.",
			},
			"toggle_on_list": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Show this environment as a toggle on the feature list page.",
			},
			"default_state": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Default enabled state for new features in this environment.",
			},
			"projects": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Project IDs this environment is scoped to. Empty means all projects.",
			},
			"parent": schema.StringAttribute{
				Optional:    true,
				Description: "An environment to inherit feature rules from (requires an enterprise license).",
			},
		},
	}
}

func (r *environmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *environmentResource) apply(state *environmentResourceModel, e *client.Environment) {
	state.ID = types.StringValue(e.ID)
	state.Description = types.StringValue(e.Description)
	state.ToggleOnList = types.BoolValue(e.ToggleOnList)
	state.DefaultState = types.BoolValue(e.DefaultState)
	state.Projects = sliceToStringList(e.Projects)
	if e.Parent == "" {
		state.Parent = types.StringNull()
	} else {
		state.Parent = types.StringValue(e.Parent)
	}
}

func (r *environmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan environmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.EnvironmentCreateInput{
		ID:           plan.ID.ValueString(),
		Description:  optString(plan.Description),
		ToggleOnList: optBool(plan.ToggleOnList),
		DefaultState: optBool(plan.DefaultState),
		Projects:     stringListToSlice(ctx, plan.Projects, &resp.Diagnostics),
		Parent:       optString(plan.Parent),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateEnvironment(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create environment", err.Error())
		return
	}
	// The create response does not echo `parent`; preserve the planned value.
	parent := plan.Parent
	r.apply(&plan, created)
	if !parent.IsNull() {
		plan.Parent = parent
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *environmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state environmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, err := r.client.GetEnvironment(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read environment", err.Error())
		return
	}
	r.apply(&state, env)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *environmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan environmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.EnvironmentUpdateInput{
		Description:  optString(plan.Description),
		ToggleOnList: optBool(plan.ToggleOnList),
		DefaultState: optBool(plan.DefaultState),
		Projects:     stringListToSlice(ctx, plan.Projects, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateEnvironment(ctx, plan.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update environment", err.Error())
		return
	}
	parent := plan.Parent
	r.apply(&plan, updated)
	if !parent.IsNull() {
		plan.Parent = parent
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *environmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state environmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteEnvironment(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete environment", err.Error())
	}
}

func (r *environmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
