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
	_ resource.Resource                = (*experimentTemplateResource)(nil)
	_ resource.ResourceWithConfigure   = (*experimentTemplateResource)(nil)
	_ resource.ResourceWithImportState = (*experimentTemplateResource)(nil)
)

func newExperimentTemplateResource() resource.Resource { return &experimentTemplateResource{} }

type experimentTemplateResource struct {
	client *client.Client
}

type experimentTemplateMetadataModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

type experimentTemplateSavedGroupModel struct {
	Match types.String `tfsdk:"match"`
	IDs   types.List   `tfsdk:"ids"`
}

type experimentTemplatePrereqModel struct {
	ID        types.String `tfsdk:"id"`
	Condition types.String `tfsdk:"condition"`
}

type experimentTemplateTargetingModel struct {
	Coverage      types.Float64                       `tfsdk:"coverage"`
	Condition     types.String                        `tfsdk:"condition"`
	SavedGroups   []experimentTemplateSavedGroupModel `tfsdk:"saved_groups"`
	Prerequisites []experimentTemplatePrereqModel     `tfsdk:"prerequisites"`
}

type experimentTemplateResourceModel struct {
	ID                     types.String                      `tfsdk:"id"`
	Project                types.String                      `tfsdk:"project"`
	Owner                  types.String                      `tfsdk:"owner"`
	OwnerEmail             types.String                      `tfsdk:"owner_email"`
	TemplateMetadata       *experimentTemplateMetadataModel  `tfsdk:"template_metadata"`
	Type                   types.String                      `tfsdk:"type"`
	Hypothesis             types.String                      `tfsdk:"hypothesis"`
	Description            types.String                      `tfsdk:"description"`
	Tags                   types.List                        `tfsdk:"tags"`
	Datasource             types.String                      `tfsdk:"datasource"`
	ExposureQueryID        types.String                      `tfsdk:"exposure_query_id"`
	HashAttribute          types.String                      `tfsdk:"hash_attribute"`
	FallbackAttribute      types.String                      `tfsdk:"fallback_attribute"`
	DisableStickyBucketing types.Bool                        `tfsdk:"disable_sticky_bucketing"`
	GoalMetrics            types.List                        `tfsdk:"goal_metrics"`
	SecondaryMetrics       types.List                        `tfsdk:"secondary_metrics"`
	GuardrailMetrics       types.List                        `tfsdk:"guardrail_metrics"`
	ActivationMetric       types.String                      `tfsdk:"activation_metric"`
	StatsEngine            types.String                      `tfsdk:"stats_engine"`
	Segment                types.String                      `tfsdk:"segment"`
	SkipPartialData        types.Bool                        `tfsdk:"skip_partial_data"`
	Targeting              *experimentTemplateTargetingModel `tfsdk:"targeting"`
	DateCreated            types.String                      `tfsdk:"date_created"`
	DateUpdated            types.String                      `tfsdk:"date_updated"`
}

func (r *experimentTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_experiment_template"
}

func (r *experimentTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook experiment template. Templates capture default experiment " +
			"settings (datasource, metrics, targeting, stats engine) that can be reused when " +
			"creating new experiments.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique experiment template identifier.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"project": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Project ID this template belongs to.",
			},
			"owner":       schema.StringAttribute{Computed: true, Description: "User ID of the template owner."},
			"owner_email": schema.StringAttribute{Computed: true, Description: "Email of the owner, when resolvable."},
			"template_metadata": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Template name and description.",
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{Required: true, Description: "Template name."},
					"description": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Template description.",
					},
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Template type. Currently only `standard` is supported.",
			},
			"hypothesis":  schema.StringAttribute{Optional: true, Computed: true, Description: "Experiment hypothesis."},
			"description": schema.StringAttribute{Optional: true, Computed: true, Description: "Experiment description."},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags applied to experiments created from this template.",
			},
			"datasource": schema.StringAttribute{
				Required:    true,
				Description: "Datasource ID used for analysis.",
			},
			"exposure_query_id": schema.StringAttribute{
				Required:    true,
				Description: "Exposure (experiment assignment) query ID within the datasource.",
			},
			"hash_attribute": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Attribute used for variation assignment/bucketing.",
			},
			"fallback_attribute": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Fallback bucketing attribute used for sticky bucketing.",
			},
			"disable_sticky_bucketing": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Disable sticky bucketing for experiments from this template.",
			},
			"goal_metrics": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Goal metric IDs.",
			},
			"secondary_metrics": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Secondary metric IDs.",
			},
			"guardrail_metrics": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Guardrail metric IDs.",
			},
			"activation_metric": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Activation metric ID.",
			},
			"stats_engine": schema.StringAttribute{
				Required:    true,
				Description: "Statistics engine: `bayesian` or `frequentist`.",
			},
			"segment": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Segment ID to restrict analysis to.",
			},
			"skip_partial_data": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Exclude in-progress conversions from analysis.",
			},
			"targeting": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Default targeting configuration for experiments from this template.",
				Attributes: map[string]schema.Attribute{
					"coverage": schema.Float64Attribute{
						Required:    true,
						Description: "Fraction of traffic included in the experiment (0-1).",
					},
					"condition": schema.StringAttribute{
						Required:    true,
						Description: "MongoDB-style targeting condition (JSON string). Use `{}` for none.",
					},
					"saved_groups": schema.ListNestedAttribute{
						Optional:     true,
						Description:  "Saved group targeting clauses.",
						NestedObject: experimentTemplateSavedGroupSchema(),
					},
					"prerequisites": schema.ListNestedAttribute{
						Optional:     true,
						Description:  "Feature prerequisites that must be satisfied.",
						NestedObject: experimentTemplatePrereqSchema(),
					},
				},
			},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

func experimentTemplateSavedGroupSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"match": schema.StringAttribute{Required: true, Description: "Match type: `all`, `any`, or `none`."},
			"ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Saved group IDs.",
			},
		},
	}
}

func experimentTemplatePrereqSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id":        schema.StringAttribute{Required: true, Description: "Prerequisite feature ID."},
			"condition": schema.StringAttribute{Required: true, Description: "Condition the prerequisite must satisfy (JSON string)."},
		},
	}
}

func (r *experimentTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (m experimentTemplateResourceModel) toInput(ctx context.Context, da *diagAppender) client.ExperimentTemplateInput {
	in := client.ExperimentTemplateInput{
		Project:                optString(m.Project),
		Type:                   m.Type.ValueString(),
		Hypothesis:             optString(m.Hypothesis),
		Description:            optString(m.Description),
		Tags:                   da.strings(ctx, m.Tags),
		Datasource:             m.Datasource.ValueString(),
		ExposureQueryID:        m.ExposureQueryID.ValueString(),
		HashAttribute:          optString(m.HashAttribute),
		FallbackAttribute:      optString(m.FallbackAttribute),
		DisableStickyBucketing: optBool(m.DisableStickyBucketing),
		GoalMetrics:            da.strings(ctx, m.GoalMetrics),
		SecondaryMetrics:       da.strings(ctx, m.SecondaryMetrics),
		GuardrailMetrics:       da.strings(ctx, m.GuardrailMetrics),
		ActivationMetric:       optString(m.ActivationMetric),
		StatsEngine:            m.StatsEngine.ValueString(),
		Segment:                optString(m.Segment),
		SkipPartialData:        optBool(m.SkipPartialData),
	}
	if m.TemplateMetadata != nil {
		in.TemplateMetadata = client.ExperimentTemplateMetadata{
			Name:        m.TemplateMetadata.Name.ValueString(),
			Description: m.TemplateMetadata.Description.ValueString(),
		}
	}
	if m.Targeting != nil {
		in.Targeting = client.ExperimentTemplateTargeting{
			Coverage:  m.Targeting.Coverage.ValueFloat64(),
			Condition: m.Targeting.Condition.ValueString(),
		}
		for _, sg := range m.Targeting.SavedGroups {
			in.Targeting.SavedGroups = append(in.Targeting.SavedGroups, client.ExperimentTemplateSavedGroup{
				Match: sg.Match.ValueString(),
				IDs:   da.strings(ctx, sg.IDs),
			})
		}
		for _, pr := range m.Targeting.Prerequisites {
			in.Targeting.Prerequisites = append(in.Targeting.Prerequisites, client.ExperimentTemplatePrerequisite{
				ID:        pr.ID.ValueString(),
				Condition: pr.Condition.ValueString(),
			})
		}
	}
	return in
}

func (r *experimentTemplateResource) apply(state *experimentTemplateResourceModel, t *client.ExperimentTemplate) {
	state.ID = types.StringValue(t.ID)
	state.Project = types.StringValue(t.Project)
	state.Owner = types.StringValue(t.Owner)
	state.OwnerEmail = types.StringValue(t.OwnerEmail)
	state.TemplateMetadata = &experimentTemplateMetadataModel{
		Name:        types.StringValue(t.TemplateMetadata.Name),
		Description: types.StringValue(t.TemplateMetadata.Description),
	}
	state.Type = types.StringValue(t.Type)
	state.Hypothesis = types.StringValue(t.Hypothesis)
	state.Description = types.StringValue(t.Description)
	state.Tags = sliceToStringList(t.Tags)
	state.Datasource = types.StringValue(t.Datasource)
	state.ExposureQueryID = types.StringValue(t.ExposureQueryID)
	state.HashAttribute = types.StringValue(t.HashAttribute)
	state.FallbackAttribute = types.StringValue(t.FallbackAttribute)
	state.DisableStickyBucketing = types.BoolValue(t.DisableStickyBucketing)
	state.GoalMetrics = sliceToStringList(t.GoalMetrics)
	state.SecondaryMetrics = sliceToStringList(t.SecondaryMetrics)
	state.GuardrailMetrics = sliceToStringList(t.GuardrailMetrics)
	state.ActivationMetric = types.StringValue(t.ActivationMetric)
	state.StatsEngine = types.StringValue(t.StatsEngine)
	state.Segment = types.StringValue(t.Segment)
	state.SkipPartialData = types.BoolValue(t.SkipPartialData)

	targeting := &experimentTemplateTargetingModel{
		Coverage:  types.Float64Value(t.Targeting.Coverage),
		Condition: types.StringValue(t.Targeting.Condition),
	}
	for _, sg := range t.Targeting.SavedGroups {
		targeting.SavedGroups = append(targeting.SavedGroups, experimentTemplateSavedGroupModel{
			Match: types.StringValue(sg.Match),
			IDs:   sliceToStringList(sg.IDs),
		})
	}
	for _, pr := range t.Targeting.Prerequisites {
		targeting.Prerequisites = append(targeting.Prerequisites, experimentTemplatePrereqModel{
			ID:        types.StringValue(pr.ID),
			Condition: types.StringValue(pr.Condition),
		})
	}
	state.Targeting = targeting

	state.DateCreated = types.StringValue(t.DateCreated)
	state.DateUpdated = types.StringValue(t.DateUpdated)
}

func (r *experimentTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan experimentTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateExperimentTemplate(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create experiment template", err.Error())
		return
	}
	r.apply(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *experimentTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state experimentTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	t, err := r.client.GetExperimentTemplate(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read experiment template", err.Error())
		return
	}
	r.apply(&state, t)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *experimentTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan experimentTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state experimentTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateExperimentTemplate(ctx, state.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update experiment template", err.Error())
		return
	}
	r.apply(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *experimentTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state experimentTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteExperimentTemplate(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete experiment template", err.Error())
	}
}

func (r *experimentTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
