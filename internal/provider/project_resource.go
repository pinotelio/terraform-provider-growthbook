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
	_ resource.Resource                = (*projectResource)(nil)
	_ resource.ResourceWithConfigure   = (*projectResource)(nil)
	_ resource.ResourceWithImportState = (*projectResource)(nil)
)

func newProjectResource() resource.Resource { return &projectResource{} }

type projectResource struct {
	client *client.Client
}

type projectResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	StatsEngine types.String `tfsdk:"stats_engine"`
	DateCreated types.String `tfsdk:"date_created"`
	DateUpdated types.String `tfsdk:"date_updated"`
}

func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook project. Projects organize feature flags, experiments, " +
			"metrics, and other resources, and back project-scoped permissions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique project identifier (e.g. `prj_...`).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable project name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional description of the project.",
			},
			"stats_engine": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Default statistics engine for the project: `bayesian` or `frequentist`.",
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

func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (m projectResourceModel) toInput() client.ProjectInput {
	in := client.ProjectInput{
		Name:        m.Name.ValueString(),
		Description: optString(m.Description),
	}
	if s := optString(m.StatsEngine); s != nil {
		in.Settings = &client.ProjectSettings{StatsEngine: *s}
	}
	return in
}

func (r *projectResource) apply(state *projectResourceModel, p *client.Project) {
	state.ID = types.StringValue(p.ID)
	state.Name = types.StringValue(p.Name)
	state.Description = types.StringValue(p.Description)
	state.StatsEngine = types.StringValue(p.Settings.StatsEngine)
	state.DateCreated = types.StringValue(p.DateCreated)
	state.DateUpdated = types.StringValue(p.DateUpdated)
}

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateProject(ctx, plan.toInput())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create project", err.Error())
		return
	}
	r.apply(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.GetProject(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read project", err.Error())
		return
	}
	r.apply(&state, project)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateProject(ctx, state.ID.ValueString(), plan.toInput())
	if err != nil {
		resp.Diagnostics.AddError("Unable to update project", err.Error())
		return
	}
	r.apply(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteProject(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete project", err.Error())
	}
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
