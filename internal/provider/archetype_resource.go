package provider

import (
	"context"
	"encoding/json"
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
	_ resource.Resource                = (*archetypeResource)(nil)
	_ resource.ResourceWithConfigure   = (*archetypeResource)(nil)
	_ resource.ResourceWithImportState = (*archetypeResource)(nil)
)

func newArchetypeResource() resource.Resource { return &archetypeResource{} }

type archetypeResource struct {
	client *client.Client
}

type archetypeResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Owner        types.String `tfsdk:"owner"`
	OwnerEmail   types.String `tfsdk:"owner_email"`
	IsPublic     types.Bool   `tfsdk:"is_public"`
	Attributes   types.String `tfsdk:"attributes"`
	Projects     types.List   `tfsdk:"projects"`
	Environments types.List   `tfsdk:"environments"`
	DateCreated  types.String `tfsdk:"date_created"`
	DateUpdated  types.String `tfsdk:"date_updated"`
}

func (r *archetypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_archetype"
}

func (r *archetypeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook archetype. Archetypes are reusable sets of user " +
			"attributes used to evaluate and debug feature flags for representative users.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique archetype identifier.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable archetype name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional description of the archetype.",
			},
			"owner": schema.StringAttribute{
				Computed:    true,
				Description: "User ID of the archetype owner (or legacy owner name/email).",
			},
			"owner_email": schema.StringAttribute{
				Computed:    true,
				Description: "Resolved email of the archetype owner, when available.",
			},
			"is_public": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the archetype is available to other team members.",
			},
			"attributes": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "JSON-encoded object of attribute values to set when using this archetype.",
			},
			"projects": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Project IDs this archetype is scoped to. Empty means all projects.",
			},
			"environments": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Environments this archetype is limited to. Empty means all environments.",
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

func (r *archetypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (m archetypeResourceModel) toInput(ctx context.Context, diags *diagAppender) client.ArchetypeInput {
	in := client.ArchetypeInput{
		Name:         m.Name.ValueString(),
		Description:  optString(m.Description),
		IsPublic:     m.IsPublic.ValueBool(),
		Attributes:   archetypeAttributesToRaw(m.Attributes),
		Projects:     diags.strings(ctx, m.Projects),
		Environments: diags.strings(ctx, m.Environments),
	}
	return in
}

func (r *archetypeResource) apply(state *archetypeResourceModel, a *client.Archetype) {
	state.ID = types.StringValue(a.ID)
	state.Name = types.StringValue(a.Name)
	state.Description = types.StringValue(a.Description)
	state.Owner = types.StringValue(a.Owner)
	state.OwnerEmail = types.StringValue(a.OwnerEmail)
	state.IsPublic = types.BoolValue(a.IsPublic)
	state.Attributes = archetypeAttributesToString(a.Attributes)
	state.Projects = sliceToStringList(a.Projects)
	state.Environments = sliceToStringList(a.Environments)
	state.DateCreated = types.StringValue(a.DateCreated)
	state.DateUpdated = types.StringValue(a.DateUpdated)
}

// archetypeAttributesToRaw converts the Terraform JSON string into raw JSON for
// the request body. A null/unknown/empty value yields nil so the field is
// omitted.
func archetypeAttributesToRaw(v types.String) json.RawMessage {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	if s == "" {
		return nil
	}
	return json.RawMessage(s)
}

// archetypeAttributesToString converts raw JSON from the API into a Terraform
// string, using null when there is no payload.
func archetypeAttributesToString(raw json.RawMessage) types.String {
	if len(raw) == 0 {
		return types.StringNull()
	}
	return types.StringValue(string(raw))
}

func (r *archetypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan archetypeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateArchetype(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create archetype", err.Error())
		return
	}
	r.apply(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *archetypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state archetypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	arch, err := r.client.GetArchetype(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read archetype", err.Error())
		return
	}
	r.apply(&state, arch)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *archetypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan archetypeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state archetypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateArchetype(ctx, state.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update archetype", err.Error())
		return
	}
	r.apply(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *archetypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state archetypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteArchetype(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete archetype", err.Error())
	}
}

func (r *archetypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
