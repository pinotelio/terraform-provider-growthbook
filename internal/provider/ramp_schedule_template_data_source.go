package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*rampScheduleTemplateDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*rampScheduleTemplateDataSource)(nil)
)

func newRampScheduleTemplateDataSource() datasource.DataSource {
	return &rampScheduleTemplateDataSource{}
}

type rampScheduleTemplateDataSource struct {
	client *client.Client
}

func (d *rampScheduleTemplateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ramp_schedule_template"
}

func (d *rampScheduleTemplateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	savedGroups := schema.ListNestedAttribute{
		Computed:    true,
		Description: "Saved group targeting clauses.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"match": schema.StringAttribute{Computed: true, Description: "Match type."},
				"ids":   schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Saved group IDs."},
			},
		},
	}
	prerequisites := schema.ListNestedAttribute{
		Computed:    true,
		Description: "Feature prerequisites.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id":        schema.StringAttribute{Computed: true, Description: "Prerequisite feature ID."},
				"condition": schema.StringAttribute{Computed: true, Description: "Prerequisite condition (JSON string)."},
			},
		},
	}

	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook ramp schedule template by ID.",
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Required: true, Description: "Unique ramp schedule template identifier."},
			"name": schema.StringAttribute{Computed: true, Description: "Template name."},
			"steps": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Ordered ramp steps.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"interval":       schema.Float64Attribute{Computed: true, Description: "Hold duration in seconds (null = no time gate)."},
						"approval_notes": schema.StringAttribute{Computed: true, Description: "Approval notes."},
						"monitored":      schema.BoolAttribute{Computed: true, Description: "Whether the step runs traffic analysis."},
						"hold_conditions": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Advancement gating conditions.",
							Attributes: map[string]schema.Attribute{
								"min_sample_size":   schema.Int64Attribute{Computed: true, Description: "Minimum sample size."},
								"requires_approval": schema.BoolAttribute{Computed: true, Description: "Whether approval is required."},
							},
						},
						"actions": schema.ListNestedAttribute{
							Computed:    true,
							Description: "Step actions.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"target_type": schema.StringAttribute{Computed: true, Description: "Target type."},
									"target_id":   schema.StringAttribute{Computed: true, Description: "Target feature ID."},
									"patch": schema.SingleNestedAttribute{
										Computed:    true,
										Description: "Feature-rule patch.",
										Attributes: map[string]schema.Attribute{
											"rule_id":          schema.StringAttribute{Computed: true, Description: "Feature rule ID."},
											"coverage":         schema.Float64Attribute{Computed: true, Description: "Traffic fraction (0-1)."},
											"condition":        schema.StringAttribute{Computed: true, Description: "Targeting condition (JSON string)."},
											"saved_groups":     savedGroups,
											"prerequisites":    prerequisites,
											"all_environments": schema.BoolAttribute{Computed: true, Description: "Apply to all environments."},
											"environments":     schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Environments."},
											"enabled":          schema.BoolAttribute{Computed: true, Description: "Enable/disable the rule."},
										},
									},
								},
							},
						},
					},
				},
			},
			"end_patch": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Patch applied when the schedule completes.",
				Attributes: map[string]schema.Attribute{
					"coverage":         schema.Float64Attribute{Computed: true, Description: "Traffic fraction (0-1)."},
					"condition":        schema.StringAttribute{Computed: true, Description: "Targeting condition (JSON string)."},
					"saved_groups":     savedGroups,
					"prerequisites":    prerequisites,
					"all_environments": schema.BoolAttribute{Computed: true, Description: "Apply to all environments."},
					"environments":     schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Environments."},
				},
			},
			"official": schema.BoolAttribute{Computed: true, Description: "Whether the template is org-official."},
			"monitoring_config": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Monitoring configuration.",
				Attributes: map[string]schema.Attribute{
					"datasource_id":                 schema.StringAttribute{Computed: true, Description: "Datasource ID."},
					"exposure_query_id":             schema.StringAttribute{Computed: true, Description: "Exposure query ID."},
					"guardrail_metric_ids":          schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Guardrail metric IDs."},
					"signal_metric_ids":             schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Signal metric IDs."},
					"update_schedule_minutes":       schema.Float64Attribute{Computed: true, Description: "Refresh interval minutes."},
					"monitoring_mode":               schema.StringAttribute{Computed: true, Description: "Monitoring mode."},
					"auto_update":                   schema.BoolAttribute{Computed: true, Description: "Auto-update analysis."},
					"srm_action":                    schema.StringAttribute{Computed: true, Description: "SRM action."},
					"no_traffic_action":             schema.StringAttribute{Computed: true, Description: "No-traffic action."},
					"no_traffic_grace_period_hours": schema.Float64Attribute{Computed: true, Description: "No-traffic grace period (hours)."},
					"multiple_exposure_action":      schema.StringAttribute{Computed: true, Description: "Multiple exposure action."},
				},
			},
			"lockdown_config": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Lockdown configuration.",
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{Computed: true, Description: "Lockdown mode."},
				},
			},
			"order":        schema.Float64Attribute{Computed: true, Description: "Display order within the org."},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

func (d *rampScheduleTemplateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *rampScheduleTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data rampScheduleTemplateResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	t, err := d.client.GetRampScheduleTemplate(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read ramp schedule template", err.Error())
		return
	}
	r := &rampScheduleTemplateResource{}
	state := r.apply(t)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
