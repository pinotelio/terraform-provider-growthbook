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
	_ resource.Resource                = (*rampScheduleTemplateResource)(nil)
	_ resource.ResourceWithConfigure   = (*rampScheduleTemplateResource)(nil)
	_ resource.ResourceWithImportState = (*rampScheduleTemplateResource)(nil)
)

func newRampScheduleTemplateResource() resource.Resource { return &rampScheduleTemplateResource{} }

type rampScheduleTemplateResource struct {
	client *client.Client
}

type rampScheduleSavedGroupModel struct {
	Match types.String `tfsdk:"match"`
	IDs   types.List   `tfsdk:"ids"`
}

type rampSchedulePrereqModel struct {
	ID        types.String `tfsdk:"id"`
	Condition types.String `tfsdk:"condition"`
}

type rampSchedulePatchModel struct {
	RuleID          types.String                  `tfsdk:"rule_id"`
	Coverage        types.Float64                 `tfsdk:"coverage"`
	Condition       types.String                  `tfsdk:"condition"`
	SavedGroups     []rampScheduleSavedGroupModel `tfsdk:"saved_groups"`
	Prerequisites   []rampSchedulePrereqModel     `tfsdk:"prerequisites"`
	AllEnvironments types.Bool                    `tfsdk:"all_environments"`
	Environments    types.List                    `tfsdk:"environments"`
	Enabled         types.Bool                    `tfsdk:"enabled"`
}

type rampScheduleActionModel struct {
	TargetType types.String            `tfsdk:"target_type"`
	TargetID   types.String            `tfsdk:"target_id"`
	Patch      *rampSchedulePatchModel `tfsdk:"patch"`
}

type rampScheduleHoldConditionsModel struct {
	MinSampleSize    types.Int64 `tfsdk:"min_sample_size"`
	RequiresApproval types.Bool  `tfsdk:"requires_approval"`
}

type rampScheduleStepModel struct {
	Interval       types.Float64                    `tfsdk:"interval"`
	ApprovalNotes  types.String                     `tfsdk:"approval_notes"`
	Monitored      types.Bool                       `tfsdk:"monitored"`
	HoldConditions *rampScheduleHoldConditionsModel `tfsdk:"hold_conditions"`
	Actions        []rampScheduleActionModel        `tfsdk:"actions"`
}

type rampScheduleEndPatchModel struct {
	Coverage        types.Float64                 `tfsdk:"coverage"`
	Condition       types.String                  `tfsdk:"condition"`
	SavedGroups     []rampScheduleSavedGroupModel `tfsdk:"saved_groups"`
	Prerequisites   []rampSchedulePrereqModel     `tfsdk:"prerequisites"`
	AllEnvironments types.Bool                    `tfsdk:"all_environments"`
	Environments    types.List                    `tfsdk:"environments"`
}

type rampScheduleMonitoringConfigModel struct {
	DatasourceID              types.String  `tfsdk:"datasource_id"`
	ExposureQueryID           types.String  `tfsdk:"exposure_query_id"`
	GuardrailMetricIDs        types.List    `tfsdk:"guardrail_metric_ids"`
	SignalMetricIDs           types.List    `tfsdk:"signal_metric_ids"`
	UpdateScheduleMinutes     types.Float64 `tfsdk:"update_schedule_minutes"`
	MonitoringMode            types.String  `tfsdk:"monitoring_mode"`
	AutoUpdate                types.Bool    `tfsdk:"auto_update"`
	SRMAction                 types.String  `tfsdk:"srm_action"`
	NoTrafficAction           types.String  `tfsdk:"no_traffic_action"`
	NoTrafficGracePeriodHours types.Float64 `tfsdk:"no_traffic_grace_period_hours"`
	MultipleExposureAction    types.String  `tfsdk:"multiple_exposure_action"`
}

type rampScheduleLockdownConfigModel struct {
	Mode types.String `tfsdk:"mode"`
}

type rampScheduleTemplateResourceModel struct {
	ID               types.String                       `tfsdk:"id"`
	Name             types.String                       `tfsdk:"name"`
	Steps            []rampScheduleStepModel            `tfsdk:"steps"`
	EndPatch         *rampScheduleEndPatchModel         `tfsdk:"end_patch"`
	Official         types.Bool                         `tfsdk:"official"`
	MonitoringConfig *rampScheduleMonitoringConfigModel `tfsdk:"monitoring_config"`
	LockdownConfig   *rampScheduleLockdownConfigModel   `tfsdk:"lockdown_config"`
	Order            types.Float64                      `tfsdk:"order"`
	DateCreated      types.String                       `tfsdk:"date_created"`
	DateUpdated      types.String                       `tfsdk:"date_updated"`
}

func (r *rampScheduleTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ramp_schedule_template"
}

func (r *rampScheduleTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook ramp schedule template. Ramp schedules define a sequence of " +
			"steps that progressively adjust a feature rule's coverage/targeting, with optional " +
			"hold gates and monitoring.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique ramp schedule template identifier.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the ramp schedule template.",
			},
			"steps": schema.ListNestedAttribute{
				Required:     true,
				Description:  "Ordered list of ramp steps. Write-only: not refreshed from the server.",
				NestedObject: rampScheduleStepSchema(),
			},
			"end_patch": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Patch applied to the feature rule when the schedule completes. Write-only: not refreshed from the server.",
				Attributes:  rampSchedulePatchBaseAttributes(false),
			},
			"official": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether this is an organization-official template.",
			},
			"monitoring_config": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Automated monitoring configuration for the rollout. Write-only: not refreshed from the server.",
				Attributes: map[string]schema.Attribute{
					"datasource_id":     schema.StringAttribute{Required: true, Description: "Datasource ID for monitoring."},
					"exposure_query_id": schema.StringAttribute{Required: true, Description: "Exposure query ID."},
					"guardrail_metric_ids": schema.ListAttribute{
						Required:    true,
						ElementType: types.StringType,
						Description: "Guardrail metric IDs (at least one).",
					},
					"signal_metric_ids": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "Signal metric IDs.",
					},
					"update_schedule_minutes":       schema.Float64Attribute{Optional: true, Description: "Monitoring refresh interval in minutes (>= 10)."},
					"monitoring_mode":               schema.StringAttribute{Optional: true, Description: "Monitoring mode: `auto` or `manual`."},
					"auto_update":                   schema.BoolAttribute{Optional: true, Description: "Automatically update analysis."},
					"srm_action":                    schema.StringAttribute{Optional: true, Description: "Action on SRM detection: `rollback`, `hold`, or `warn`."},
					"no_traffic_action":             schema.StringAttribute{Optional: true, Description: "Action when no traffic: `rollback`, `hold`, or `warn`."},
					"no_traffic_grace_period_hours": schema.Float64Attribute{Optional: true, Description: "Hours to wait for traffic before applying the no-traffic action."},
					"multiple_exposure_action":      schema.StringAttribute{Optional: true, Description: "Action on multiple exposures: `rollback`, `hold`, or `warn`."},
				},
			},
			"lockdown_config": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Lockdown configuration. Write-only: not refreshed from the server.",
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{Required: true, Description: "Lockdown mode: `none` or `locked`."},
				},
			},
			"order": schema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Display order within the org. Omit to append to the end.",
			},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

func rampScheduleStepSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"interval": schema.Float64Attribute{
				Optional:    true,
				Description: "Hold duration in seconds before this step's gates are evaluated. Null means no time gate.",
			},
			"approval_notes": schema.StringAttribute{Optional: true, Description: "Notes shown during approval."},
			"monitored":      schema.BoolAttribute{Optional: true, Description: "Run A/B traffic analysis while this step is active."},
			"hold_conditions": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Conditions gating advancement to the next step.",
				Attributes: map[string]schema.Attribute{
					"min_sample_size":   schema.Int64Attribute{Optional: true, Description: "Minimum sample size before advancing."},
					"requires_approval": schema.BoolAttribute{Optional: true, Description: "Require manual approval before advancing."},
				},
			},
			"actions": schema.ListNestedAttribute{
				Required:     true,
				Description:  "Actions applied at this step.",
				NestedObject: rampScheduleActionSchema(),
			},
		},
	}
}

func rampScheduleActionSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"target_type": schema.StringAttribute{Required: true, Description: "Target type. Currently only `feature-rule`."},
			"target_id":   schema.StringAttribute{Required: true, Description: "ID of the targeted feature."},
			"patch": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Feature-rule patch applied at this step.",
				Attributes:  rampSchedulePatchBaseAttributes(true),
			},
		},
	}
}

// rampSchedulePatchBaseAttributes returns the shared patch attributes. When
// ruleScoped is true the `rule_id` and `enabled` fields (step patches only) are
// included; end patches omit them.
func rampSchedulePatchBaseAttributes(ruleScoped bool) map[string]schema.Attribute {
	attrs := map[string]schema.Attribute{
		"coverage":  schema.Float64Attribute{Optional: true, Description: "Traffic fraction (0-1)."},
		"condition": schema.StringAttribute{Optional: true, Description: "Targeting condition (JSON string)."},
		"saved_groups": schema.ListNestedAttribute{
			Optional:     true,
			Description:  "Saved group targeting clauses.",
			NestedObject: rampScheduleSavedGroupSchema(),
		},
		"prerequisites": schema.ListNestedAttribute{
			Optional:     true,
			Description:  "Feature prerequisites.",
			NestedObject: rampSchedulePrereqSchema(),
		},
		"all_environments": schema.BoolAttribute{Optional: true, Description: "Apply to all environments."},
		"environments": schema.ListAttribute{
			Optional:    true,
			ElementType: types.StringType,
			Description: "Environments to apply to (when not all).",
		},
	}
	if ruleScoped {
		attrs["rule_id"] = schema.StringAttribute{Required: true, Description: "ID of the feature rule to patch."}
		attrs["enabled"] = schema.BoolAttribute{Optional: true, Description: "Enable or disable the rule."}
	}
	return attrs
}

func rampScheduleSavedGroupSchema() schema.NestedAttributeObject {
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

func rampSchedulePrereqSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id":        schema.StringAttribute{Required: true, Description: "Prerequisite feature ID."},
			"condition": schema.StringAttribute{Required: true, Description: "Prerequisite condition (JSON string)."},
		},
	}
}

func (r *rampScheduleTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (m rampScheduleTemplateResourceModel) toInput(ctx context.Context, da *diagAppender) client.RampScheduleTemplateInput {
	in := client.RampScheduleTemplateInput{
		Name:     m.Name.ValueString(),
		Official: optBool(m.Official),
		Order:    optFloat64(m.Order),
	}
	for _, s := range m.Steps {
		step := client.RampScheduleStep{
			Interval:      optFloat64(s.Interval),
			ApprovalNotes: optString(s.ApprovalNotes),
			Monitored:     optBool(s.Monitored),
		}
		if s.HoldConditions != nil {
			step.HoldConditions = &client.RampScheduleHoldConditions{
				MinSampleSize:    optInt64(s.HoldConditions.MinSampleSize),
				RequiresApproval: optBool(s.HoldConditions.RequiresApproval),
			}
		}
		for _, a := range s.Actions {
			action := client.RampScheduleAction{
				TargetType: a.TargetType.ValueString(),
				TargetID:   a.TargetID.ValueString(),
			}
			if a.Patch != nil {
				action.Patch = client.RampSchedulePatch{
					RuleID:          a.Patch.RuleID.ValueString(),
					Coverage:        optFloat64(a.Patch.Coverage),
					Condition:       optString(a.Patch.Condition),
					SavedGroups:     rampScheduleSavedGroupsToInput(ctx, a.Patch.SavedGroups, da),
					Prerequisites:   rampSchedulePrereqsToInput(a.Patch.Prerequisites),
					AllEnvironments: optBool(a.Patch.AllEnvironments),
					Environments:    da.strings(ctx, a.Patch.Environments),
					Enabled:         optBool(a.Patch.Enabled),
				}
			}
			step.Actions = append(step.Actions, action)
		}
		in.Steps = append(in.Steps, step)
	}
	if m.EndPatch != nil {
		in.EndPatch = &client.RampScheduleEndPatch{
			Coverage:        optFloat64(m.EndPatch.Coverage),
			Condition:       optString(m.EndPatch.Condition),
			SavedGroups:     rampScheduleSavedGroupsToInput(ctx, m.EndPatch.SavedGroups, da),
			Prerequisites:   rampSchedulePrereqsToInput(m.EndPatch.Prerequisites),
			AllEnvironments: optBool(m.EndPatch.AllEnvironments),
			Environments:    da.strings(ctx, m.EndPatch.Environments),
		}
	}
	if m.MonitoringConfig != nil {
		in.MonitoringConfig = &client.RampScheduleMonitoringConfig{
			DatasourceID:              m.MonitoringConfig.DatasourceID.ValueString(),
			ExposureQueryID:           m.MonitoringConfig.ExposureQueryID.ValueString(),
			GuardrailMetricIDs:        da.strings(ctx, m.MonitoringConfig.GuardrailMetricIDs),
			SignalMetricIDs:           da.strings(ctx, m.MonitoringConfig.SignalMetricIDs),
			UpdateScheduleMinutes:     optFloat64(m.MonitoringConfig.UpdateScheduleMinutes),
			MonitoringMode:            m.MonitoringConfig.MonitoringMode.ValueString(),
			AutoUpdate:                optBool(m.MonitoringConfig.AutoUpdate),
			SRMAction:                 m.MonitoringConfig.SRMAction.ValueString(),
			NoTrafficAction:           m.MonitoringConfig.NoTrafficAction.ValueString(),
			NoTrafficGracePeriodHours: optFloat64(m.MonitoringConfig.NoTrafficGracePeriodHours),
			MultipleExposureAction:    m.MonitoringConfig.MultipleExposureAction.ValueString(),
		}
	}
	if m.LockdownConfig != nil {
		in.LockdownConfig = &client.RampScheduleLockdownConfig{
			Mode: m.LockdownConfig.Mode.ValueString(),
		}
	}
	return in
}

func rampScheduleSavedGroupsToInput(ctx context.Context, models []rampScheduleSavedGroupModel, da *diagAppender) []client.RampScheduleSavedGroup {
	var out []client.RampScheduleSavedGroup
	for _, sg := range models {
		out = append(out, client.RampScheduleSavedGroup{
			Match: sg.Match.ValueString(),
			IDs:   da.strings(ctx, sg.IDs),
		})
	}
	return out
}

func rampSchedulePrereqsToInput(models []rampSchedulePrereqModel) []client.RampSchedulePrerequisite {
	var out []client.RampSchedulePrerequisite
	for _, pr := range models {
		out = append(out, client.RampSchedulePrerequisite{
			ID:        pr.ID.ValueString(),
			Condition: pr.Condition.ValueString(),
		})
	}
	return out
}

func (r *rampScheduleTemplateResource) apply(t *client.RampScheduleTemplate) rampScheduleTemplateResourceModel {
	// The nested blocks (steps, end_patch, monitoring_config, lockdown_config) are
	// write-only: not refreshed from the server. The server returns
	// defaulted/normalized values that would otherwise cause "inconsistent result
	// after apply" or perpetual diffs, so the caller preserves them from the plan
	// (Create/Update) or prior state (Read) instead of populating them here.
	return rampScheduleTemplateResourceModel{
		ID:          types.StringValue(t.ID),
		Name:        types.StringValue(t.Name),
		Official:    boolPtrValue(t.Official),
		Order:       float64PtrValue(t.Order),
		DateCreated: types.StringValue(t.DateCreated),
		DateUpdated: types.StringValue(t.DateUpdated),
	}
}

func (r *rampScheduleTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan rampScheduleTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateRampScheduleTemplate(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create ramp schedule template", err.Error())
		return
	}
	state := r.apply(created)
	// Preserve write-only nested blocks from the plan (config), since apply does
	// not refresh them from the server.
	state.Steps = plan.Steps
	state.EndPatch = plan.EndPatch
	state.MonitoringConfig = plan.MonitoringConfig
	state.LockdownConfig = plan.LockdownConfig
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *rampScheduleTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state rampScheduleTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	t, err := r.client.GetRampScheduleTemplate(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read ramp schedule template", err.Error())
		return
	}
	newState := r.apply(t)
	// Preserve write-only nested blocks from prior state, since apply does not
	// refresh them from the server.
	newState.Steps = state.Steps
	newState.EndPatch = state.EndPatch
	newState.MonitoringConfig = state.MonitoringConfig
	newState.LockdownConfig = state.LockdownConfig
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *rampScheduleTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan rampScheduleTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state rampScheduleTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateRampScheduleTemplate(ctx, state.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update ramp schedule template", err.Error())
		return
	}
	newState := r.apply(updated)
	// Preserve write-only nested blocks from the plan (config), since apply does
	// not refresh them from the server.
	newState.Steps = plan.Steps
	newState.EndPatch = plan.EndPatch
	newState.MonitoringConfig = plan.MonitoringConfig
	newState.LockdownConfig = plan.LockdownConfig
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *rampScheduleTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state rampScheduleTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRampScheduleTemplate(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete ramp schedule template", err.Error())
	}
}

func (r *rampScheduleTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
