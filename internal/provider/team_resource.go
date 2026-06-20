package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ resource.Resource                = (*teamResource)(nil)
	_ resource.ResourceWithConfigure   = (*teamResource)(nil)
	_ resource.ResourceWithImportState = (*teamResource)(nil)
)

func newTeamResource() resource.Resource { return &teamResource{} }

type teamResource struct {
	client *client.Client
}

type teamResourceModel struct {
	ID                       types.String           `tfsdk:"id"`
	Name                     types.String           `tfsdk:"name"`
	Description              types.String           `tfsdk:"description"`
	Role                     types.String           `tfsdk:"role"`
	LimitAccessByEnvironment types.Bool             `tfsdk:"limit_access_by_environment"`
	Environments             types.List             `tfsdk:"environments"`
	ProjectRoles             []teamProjectRoleModel `tfsdk:"project_roles"`
	Members                  types.List             `tfsdk:"members"`
	DefaultProject           types.String           `tfsdk:"default_project"`
	CreatedBy                types.String           `tfsdk:"created_by"`
	ManagedByIdp             types.Bool             `tfsdk:"managed_by_idp"`
	DateCreated              types.String           `tfsdk:"date_created"`
	DateUpdated              types.String           `tfsdk:"date_updated"`
}

type teamProjectRoleModel struct {
	Project                  types.String `tfsdk:"project"`
	Role                     types.String `tfsdk:"role"`
	LimitAccessByEnvironment types.Bool   `tfsdk:"limit_access_by_environment"`
	Environments             types.List   `tfsdk:"environments"`
	Teams                    types.List   `tfsdk:"teams"`
}

func (r *teamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *teamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook team. Teams group members under a shared global role and " +
			"optional per-project role overrides.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique team identifier.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable team name.",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description of the team.",
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "Global role granted to members of this team (e.g. `readonly`, `engineer`, `admin`).",
			},
			"limit_access_by_environment": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the global role's environment-scoped permissions are limited to `environments`.",
			},
			"environments": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Environments the global role is limited to. Empty means all environments.",
			},
			"project_roles": schema.ListNestedAttribute{
				Optional:     true,
				Description:  "Per-project role overrides for this team.",
				NestedObject: teamProjectRoleSchema(),
			},
			"members": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "User IDs that belong to this team. When set, the provider reconciles " +
					"membership to match exactly. Leave unset to track membership without managing it.",
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"default_project": schema.StringAttribute{
				Optional:    true,
				Description: "Default project ID applied for new resources created by team members.",
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "User ID that created the team.",
			},
			"managed_by_idp": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether team membership is managed by an external identity provider.",
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

func teamProjectRoleSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"project": schema.StringAttribute{
				Required:    true,
				Description: "Project ID the role override applies to.",
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "Role granted within the project.",
			},
			"limit_access_by_environment": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the override's environment-scoped permissions are limited to `environments`.",
			},
			"environments": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Environments the override is limited to. Empty means all environments.",
			},
			"teams": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Nested team IDs associated with this project role.",
			},
		},
	}
}

func (r *teamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (m teamResourceModel) toInput(ctx context.Context, diags *diagAppender) client.TeamInput {
	in := client.TeamInput{
		Name:                     m.Name.ValueString(),
		Description:              m.Description.ValueString(),
		Role:                     m.Role.ValueString(),
		LimitAccessByEnvironment: optBool(m.LimitAccessByEnvironment),
		Environments:             diags.strings(ctx, m.Environments),
		DefaultProject:           optString(m.DefaultProject),
	}
	for _, pr := range m.ProjectRoles {
		envs := diags.strings(ctx, pr.Environments)
		if envs == nil {
			envs = []string{}
		}
		in.ProjectRoles = append(in.ProjectRoles, client.TeamProjectRole{
			Project:                  pr.Project.ValueString(),
			Role:                     pr.Role.ValueString(),
			LimitAccessByEnvironment: pr.LimitAccessByEnvironment.ValueBool(),
			Environments:             envs,
			Teams:                    diags.strings(ctx, pr.Teams),
		})
	}
	return in
}

func (r *teamResource) apply(state *teamResourceModel, t *client.Team) {
	state.ID = types.StringValue(t.ID)
	state.Name = types.StringValue(t.Name)
	state.Description = types.StringValue(t.Description)
	state.Role = types.StringValue(t.Role)
	state.LimitAccessByEnvironment = types.BoolValue(t.LimitAccessByEnvironment)
	state.Environments = sliceToStringList(t.Environments)
	state.ProjectRoles = flattenTeamProjectRoles(t.ProjectRoles)
	state.Members = sliceToStringList(t.Members)
	state.DefaultProject = types.StringValue(t.DefaultProject)
	state.CreatedBy = types.StringValue(t.CreatedBy)
	state.ManagedByIdp = types.BoolValue(t.ManagedByIdp)
	state.DateCreated = types.StringValue(t.DateCreated)
	state.DateUpdated = types.StringValue(t.DateUpdated)
}

func flattenTeamProjectRoles(roles []client.TeamProjectRole) []teamProjectRoleModel {
	if len(roles) == 0 {
		return nil
	}
	out := make([]teamProjectRoleModel, 0, len(roles))
	for _, pr := range roles {
		out = append(out, teamProjectRoleModel{
			Project:                  types.StringValue(pr.Project),
			Role:                     types.StringValue(pr.Role),
			LimitAccessByEnvironment: types.BoolValue(pr.LimitAccessByEnvironment),
			Environments:             sliceToStringList(pr.Environments),
			Teams:                    sliceToStringList(pr.Teams),
		})
	}
	return out
}

// teamMemberDiff returns the members to add and remove to move from current to
// the desired set.
func teamMemberDiff(desired, current []string) (toAdd, toRemove []string) {
	cur := make(map[string]struct{}, len(current))
	for _, m := range current {
		cur[m] = struct{}{}
	}
	des := make(map[string]struct{}, len(desired))
	for _, m := range desired {
		des[m] = struct{}{}
	}
	for _, m := range desired {
		if _, ok := cur[m]; !ok {
			toAdd = append(toAdd, m)
		}
	}
	for _, m := range current {
		if _, ok := des[m]; !ok {
			toRemove = append(toRemove, m)
		}
	}
	return toAdd, toRemove
}

func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateTeam(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create team", err.Error())
		return
	}

	// Reconcile members against the configured set, if any were provided.
	if !plan.Members.IsNull() && !plan.Members.IsUnknown() {
		desired := da.strings(ctx, plan.Members)
		if resp.Diagnostics.HasError() {
			return
		}
		if err := r.client.AddTeamMembers(ctx, created.ID, desired); err != nil {
			resp.Diagnostics.AddError("Unable to add team members", err.Error())
			return
		}
	}

	final, err := r.client.GetTeam(ctx, created.ID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read team after create", err.Error())
		return
	}
	r.apply(&plan, final)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	team, err := r.client.GetTeam(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read team", err.Error())
		return
	}
	r.apply(&state, team)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan teamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.UpdateTeam(ctx, state.ID.ValueString(), in); err != nil {
		resp.Diagnostics.AddError("Unable to update team", err.Error())
		return
	}

	// Reconcile membership unless members are unmanaged (unknown plan value).
	if !plan.Members.IsUnknown() {
		desired := da.strings(ctx, plan.Members)
		current := da.strings(ctx, state.Members)
		if resp.Diagnostics.HasError() {
			return
		}
		toAdd, toRemove := teamMemberDiff(desired, current)
		if err := r.client.AddTeamMembers(ctx, state.ID.ValueString(), toAdd); err != nil {
			resp.Diagnostics.AddError("Unable to add team members", err.Error())
			return
		}
		if err := r.client.RemoveTeamMembers(ctx, state.ID.ValueString(), toRemove); err != nil {
			resp.Diagnostics.AddError("Unable to remove team members", err.Error())
			return
		}
	}

	final, err := r.client.GetTeam(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read team after update", err.Error())
		return
	}
	r.apply(&plan, final)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteTeam(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete team", err.Error())
	}
}

func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
