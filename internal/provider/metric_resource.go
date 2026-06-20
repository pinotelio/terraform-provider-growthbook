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
	_ resource.Resource                = (*metricResource)(nil)
	_ resource.ResourceWithConfigure   = (*metricResource)(nil)
	_ resource.ResourceWithImportState = (*metricResource)(nil)
)

func newMetricResource() resource.Resource { return &metricResource{} }

type metricResource struct {
	client *client.Client
}

type metricResourceModel struct {
	ID           types.String         `tfsdk:"id"`
	Name         types.String         `tfsdk:"name"`
	Description  types.String         `tfsdk:"description"`
	Owner        types.String         `tfsdk:"owner"`
	Type         types.String         `tfsdk:"type"`
	DatasourceID types.String         `tfsdk:"datasource_id"`
	ManagedBy    types.String         `tfsdk:"managed_by"`
	Projects     types.List           `tfsdk:"projects"`
	Tags         types.List           `tfsdk:"tags"`
	Archived     types.Bool           `tfsdk:"archived"`
	Behavior     *metricBehaviorModel `tfsdk:"behavior"`
	SQL          *metricSQLModel      `tfsdk:"sql"`
	DateCreated  types.String         `tfsdk:"date_created"`
	DateUpdated  types.String         `tfsdk:"date_updated"`
}

type metricBehaviorModel struct {
	Goal                 types.String                `tfsdk:"goal"`
	CappingSettings      *metricCappingSettingsModel `tfsdk:"capping_settings"`
	WindowSettings       *metricWindowSettingsModel  `tfsdk:"window_settings"`
	PriorSettings        *metricPriorSettingsModel   `tfsdk:"prior_settings"`
	RiskThresholdSuccess types.Float64               `tfsdk:"risk_threshold_success"`
	RiskThresholdDanger  types.Float64               `tfsdk:"risk_threshold_danger"`
	MinPercentChange     types.Float64               `tfsdk:"min_percent_change"`
	MaxPercentChange     types.Float64               `tfsdk:"max_percent_change"`
	MinSampleSize        types.Float64               `tfsdk:"min_sample_size"`
	TargetMDE            types.Float64               `tfsdk:"target_mde"`
}

type metricCappingSettingsModel struct {
	Type        types.String  `tfsdk:"type"`
	Value       types.Float64 `tfsdk:"value"`
	IgnoreZeros types.Bool    `tfsdk:"ignore_zeros"`
}

type metricWindowSettingsModel struct {
	Type        types.String  `tfsdk:"type"`
	DelayValue  types.Float64 `tfsdk:"delay_value"`
	DelayUnit   types.String  `tfsdk:"delay_unit"`
	WindowValue types.Float64 `tfsdk:"window_value"`
	WindowUnit  types.String  `tfsdk:"window_unit"`
}

type metricPriorSettingsModel struct {
	Override types.Bool    `tfsdk:"override"`
	Proper   types.Bool    `tfsdk:"proper"`
	Mean     types.Float64 `tfsdk:"mean"`
	Stddev   types.Float64 `tfsdk:"stddev"`
}

type metricSQLModel struct {
	IdentifierTypes     types.List   `tfsdk:"identifier_types"`
	ConversionSQL       types.String `tfsdk:"conversion_sql"`
	UserAggregationSQL  types.String `tfsdk:"user_aggregation_sql"`
	DenominatorMetricID types.String `tfsdk:"denominator_metric_id"`
}

func (r *metricResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metric"
}

func metricOptionalFloat(desc string) schema.Float64Attribute {
	return schema.Float64Attribute{Optional: true, Description: desc}
}

func (r *metricResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook (legacy) metric. Metrics define how experiment success is " +
			"measured against a data source via a SQL query and analysis behavior settings. " +
			"The `sqlBuilder` and `mixpanel` query styles are not managed by this resource; " +
			"use the `sql` block to define the metric query.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique metric identifier (e.g. `met_...`).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable metric name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional description of the metric.",
			},
			"owner": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "User ID or email of the metric owner.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Metric type: `binomial`, `count`, `duration`, or `revenue`.",
			},
			"datasource_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the data source this metric queries. Changing this forces a new metric.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"managed_by": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Where the metric is managed from: empty (anywhere) or `api`.",
			},
			"projects": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Project IDs that can access this metric.",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags applied to the metric.",
			},
			"archived": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the metric is archived.",
			},
			"behavior": schema.SingleNestedAttribute{
				Optional: true,
				Description: "Analysis behavior settings for the metric (goal direction, capping, conversion window, bayesian " +
					"prior, thresholds). Server-applied defaults are not read back into state; only the values you " +
					"configure are tracked.",
				Attributes: map[string]schema.Attribute{
					"goal": schema.StringAttribute{
						Optional:    true,
						Description: "Metric goal direction: `increase` or `decrease`.",
					},
					"capping_settings": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Controls how metric outliers are capped.",
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Optional:    true,
								Description: "Capping type: `none`, `absolute`, or `percentile`.",
							},
							"value": metricOptionalFloat("Absolute cap value, or percentile (0.0-1.0) depending on `type`."),
							"ignore_zeros": schema.BoolAttribute{
								Optional:    true,
								Description: "When percentile capping, ignore zero values when computing the percentile.",
							},
						},
					},
					"window_settings": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Controls the conversion window for the metric.",
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Optional:    true,
								Description: "Window type: `none`, `conversion`, or `lookback`.",
							},
							"delay_value": metricOptionalFloat("How long after exposure before counting conversions."),
							"delay_unit": schema.StringAttribute{
								Optional:    true,
								Description: "Unit for `delay_value`: `minutes`, `hours`, `days`, or `weeks`.",
							},
							"window_value": metricOptionalFloat("Length of the conversion window."),
							"window_unit": schema.StringAttribute{
								Optional:    true,
								Description: "Unit for `window_value`: `minutes`, `hours`, `days`, or `weeks`.",
							},
						},
					},
					"prior_settings": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Controls the bayesian prior for the metric.",
						Attributes: map[string]schema.Attribute{
							"override": schema.BoolAttribute{
								Optional:    true,
								Description: "If false, organization default prior settings are used.",
							},
							"proper": schema.BoolAttribute{
								Optional:    true,
								Description: "If true, use the `mean` and `stddev`; otherwise use an improper flat prior.",
							},
							"mean":   metricOptionalFloat("Mean of the prior distribution of relative effects (proportion)."),
							"stddev": metricOptionalFloat("Standard deviation of the prior distribution (must be > 0)."),
						},
					},
					"risk_threshold_success": metricOptionalFloat("Deprecated risk success threshold (proportion)."),
					"risk_threshold_danger":  metricOptionalFloat("Deprecated risk danger threshold (proportion)."),
					"min_percent_change":     metricOptionalFloat("Minimum percent change to consider uplift significant (proportion)."),
					"max_percent_change":     metricOptionalFloat("Maximum percent change to consider uplift significant (proportion)."),
					"min_sample_size":        metricOptionalFloat("Minimum sample size before showing results."),
					"target_mde":             metricOptionalFloat("Target minimum detectable effect, as a proportion."),
				},
			},
			"sql": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "SQL definition of the metric. Preferred over the unsupported `sqlBuilder`/`mixpanel` styles.",
				Attributes: map[string]schema.Attribute{
					"identifier_types": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "Identifier types this metric supports.",
					},
					"conversion_sql": schema.StringAttribute{
						Optional:    true,
						Description: "SQL query returning conversion rows for the metric.",
					},
					"user_aggregation_sql": schema.StringAttribute{
						Optional:    true,
						Description: "Custom user-level aggregation SQL (default `SUM(value)`).",
					},
					"denominator_metric_id": schema.StringAttribute{
						Optional:    true,
						Description: "Denominator metric ID for funnel/ratio metrics.",
					},
				},
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

func (r *metricResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func metricBehaviorToClient(m *metricBehaviorModel) *client.MetricBehavior {
	if m == nil {
		return nil
	}
	out := &client.MetricBehavior{
		Goal:                 optString(m.Goal),
		RiskThresholdSuccess: optFloat64(m.RiskThresholdSuccess),
		RiskThresholdDanger:  optFloat64(m.RiskThresholdDanger),
		MinPercentChange:     optFloat64(m.MinPercentChange),
		MaxPercentChange:     optFloat64(m.MaxPercentChange),
		MinSampleSize:        optFloat64(m.MinSampleSize),
		TargetMDE:            optFloat64(m.TargetMDE),
	}
	if c := m.CappingSettings; c != nil {
		out.CappingSettings = &client.MetricCappingSettings{
			Type:        optString(c.Type),
			Value:       optFloat64(c.Value),
			IgnoreZeros: optBool(c.IgnoreZeros),
		}
	}
	if w := m.WindowSettings; w != nil {
		out.WindowSettings = &client.MetricWindowSettings{
			Type:        w.Type.ValueString(),
			DelayValue:  optFloat64(w.DelayValue),
			DelayUnit:   optString(w.DelayUnit),
			WindowValue: optFloat64(w.WindowValue),
			WindowUnit:  optString(w.WindowUnit),
		}
	}
	if p := m.PriorSettings; p != nil {
		out.PriorSettings = &client.MetricPriorSettings{
			Override: p.Override.ValueBool(),
			Proper:   p.Proper.ValueBool(),
			Mean:     p.Mean.ValueFloat64(),
			Stddev:   p.Stddev.ValueFloat64(),
		}
	}
	return out
}

func metricBehaviorFromClient(b *client.MetricBehavior) *metricBehaviorModel {
	if b == nil {
		return nil
	}
	out := &metricBehaviorModel{
		Goal:                 stringPtrValue(b.Goal),
		RiskThresholdSuccess: float64PtrValue(b.RiskThresholdSuccess),
		RiskThresholdDanger:  float64PtrValue(b.RiskThresholdDanger),
		MinPercentChange:     float64PtrValue(b.MinPercentChange),
		MaxPercentChange:     float64PtrValue(b.MaxPercentChange),
		MinSampleSize:        float64PtrValue(b.MinSampleSize),
		TargetMDE:            float64PtrValue(b.TargetMDE),
	}
	if c := b.CappingSettings; c != nil {
		out.CappingSettings = &metricCappingSettingsModel{
			Type:        stringPtrValue(c.Type),
			Value:       float64PtrValue(c.Value),
			IgnoreZeros: boolPtrValue(c.IgnoreZeros),
		}
	}
	if w := b.WindowSettings; w != nil {
		out.WindowSettings = &metricWindowSettingsModel{
			Type:        types.StringValue(w.Type),
			DelayValue:  float64PtrValue(w.DelayValue),
			DelayUnit:   stringPtrValue(w.DelayUnit),
			WindowValue: float64PtrValue(w.WindowValue),
			WindowUnit:  stringPtrValue(w.WindowUnit),
		}
	}
	if p := b.PriorSettings; p != nil {
		out.PriorSettings = &metricPriorSettingsModel{
			Override: types.BoolValue(p.Override),
			Proper:   types.BoolValue(p.Proper),
			Mean:     types.Float64Value(p.Mean),
			Stddev:   types.Float64Value(p.Stddev),
		}
	}
	return out
}

func metricSQLToClient(ctx context.Context, m *metricSQLModel, diags *diagAppender) *client.MetricSQL {
	if m == nil {
		return nil
	}
	return &client.MetricSQL{
		IdentifierTypes:     diags.strings(ctx, m.IdentifierTypes),
		ConversionSQL:       m.ConversionSQL.ValueString(),
		UserAggregationSQL:  m.UserAggregationSQL.ValueString(),
		DenominatorMetricID: m.DenominatorMetricID.ValueString(),
	}
}

func metricSQLFromClient(s *client.MetricSQL) *metricSQLModel {
	if s == nil {
		return nil
	}
	return &metricSQLModel{
		IdentifierTypes:     sliceToStringList(s.IdentifierTypes),
		ConversionSQL:       types.StringValue(s.ConversionSQL),
		UserAggregationSQL:  types.StringValue(s.UserAggregationSQL),
		DenominatorMetricID: types.StringValue(s.DenominatorMetricID),
	}
}

func (m metricResourceModel) toInput(ctx context.Context, diags *diagAppender) client.MetricInput {
	return client.MetricInput{
		DatasourceID: m.DatasourceID.ValueString(),
		ManagedBy:    optString(m.ManagedBy),
		Owner:        optString(m.Owner),
		Name:         m.Name.ValueString(),
		Description:  optString(m.Description),
		Type:         m.Type.ValueString(),
		Tags:         diags.strings(ctx, m.Tags),
		Projects:     diags.strings(ctx, m.Projects),
		Archived:     optBool(m.Archived),
		Behavior:     metricBehaviorToClient(m.Behavior),
		SQL:          metricSQLToClient(ctx, m.SQL, diags),
	}
}

// apply maps an API metric onto the model. The behavior and sql blocks are
// deliberately left untouched: the server applies its own defaults to those
// nested objects, so reading them back would surface perpetual drift against
// the user's (possibly partial) configuration. They are preserved exactly as
// configured instead.
func (r *metricResource) apply(state *metricResourceModel, mt *client.Metric) {
	state.ID = types.StringValue(mt.ID)
	state.Name = types.StringValue(mt.Name)
	state.Description = types.StringValue(mt.Description)
	state.Owner = types.StringValue(mt.Owner)
	state.Type = types.StringValue(mt.Type)
	state.DatasourceID = types.StringValue(mt.DatasourceID)
	state.ManagedBy = types.StringValue(mt.ManagedBy)
	state.Projects = sliceToStringList(mt.Projects)
	state.Tags = sliceToStringList(mt.Tags)
	state.Archived = types.BoolValue(mt.Archived)
	state.DateCreated = types.StringValue(mt.DateCreated)
	state.DateUpdated = types.StringValue(mt.DateUpdated)
}

func (r *metricResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan metricResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateMetric(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create metric", err.Error())
		return
	}
	r.apply(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *metricResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state metricResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	mt, err := r.client.GetMetric(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read metric", err.Error())
		return
	}
	r.apply(&state, mt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *metricResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan metricResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state metricResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateMetric(ctx, state.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update metric", err.Error())
		return
	}
	r.apply(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *metricResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state metricResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteMetric(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete metric", err.Error())
	}
}

func (r *metricResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
