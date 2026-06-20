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
	_ resource.Resource                = (*metricGroupResource)(nil)
	_ resource.ResourceWithConfigure   = (*metricGroupResource)(nil)
	_ resource.ResourceWithImportState = (*metricGroupResource)(nil)
)

func newMetricGroupResource() resource.Resource { return &metricGroupResource{} }

type metricGroupResource struct {
	client *client.Client
}

type metricGroupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Owner       types.String `tfsdk:"owner"`
	Description types.String `tfsdk:"description"`
	Tags        types.List   `tfsdk:"tags"`
	Projects    types.List   `tfsdk:"projects"`
	Metrics     types.List   `tfsdk:"metrics"`
	Datasource  types.String `tfsdk:"datasource"`
	Archived    types.Bool   `tfsdk:"archived"`
	DateCreated types.String `tfsdk:"date_created"`
	DateUpdated types.String `tfsdk:"date_updated"`
}

func (r *metricGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metric_group"
}

func (r *metricGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook metric group. Metric groups bundle an ordered set of " +
			"metrics from a single data source so they can be reused together across " +
			"experiments.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique metric group identifier (e.g. `mg_...`).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable metric group name.",
			},
			"owner": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "User ID or email of the metric group owner.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional description of the metric group.",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags applied to the metric group.",
			},
			"projects": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Project IDs the metric group is scoped to.",
			},
			"metrics": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Ordered list of metric IDs that belong to the group.",
			},
			"datasource": schema.StringAttribute{
				Required:    true,
				Description: "ID of the data source backing the metric group.",
			},
			"archived": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the metric group is archived.",
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

func (r *metricGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (m metricGroupResourceModel) toInput(ctx context.Context, diags *diagAppender) client.MetricGroupInput {
	return client.MetricGroupInput{
		Name:        m.Name.ValueString(),
		Owner:       optString(m.Owner),
		Description: m.Description.ValueString(),
		Tags:        diags.strings(ctx, m.Tags),
		Projects:    diags.strings(ctx, m.Projects),
		Metrics:     diags.strings(ctx, m.Metrics),
		Datasource:  m.Datasource.ValueString(),
		Archived:    optBool(m.Archived),
	}
}

func (r *metricGroupResource) apply(state *metricGroupResourceModel, g *client.MetricGroup) {
	state.ID = types.StringValue(g.ID)
	state.Name = types.StringValue(g.Name)
	state.Owner = types.StringValue(g.Owner)
	state.Description = types.StringValue(g.Description)
	state.Tags = sliceToStringList(g.Tags)
	state.Projects = sliceToStringList(g.Projects)
	state.Metrics = sliceToStringList(g.Metrics)
	state.Datasource = types.StringValue(g.Datasource)
	state.Archived = types.BoolValue(g.Archived)
	state.DateCreated = types.StringValue(g.DateCreated)
	state.DateUpdated = types.StringValue(g.DateUpdated)
}

func (r *metricGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan metricGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateMetricGroup(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create metric group", err.Error())
		return
	}
	r.apply(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *metricGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state metricGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	group, err := r.client.GetMetricGroup(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read metric group", err.Error())
		return
	}
	r.apply(&state, group)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *metricGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan metricGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state metricGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateMetricGroup(ctx, state.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update metric group", err.Error())
		return
	}
	r.apply(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *metricGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state metricGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteMetricGroup(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete metric group", err.Error())
	}
}

func (r *metricGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
