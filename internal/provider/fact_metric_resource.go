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
	_ resource.Resource                = (*factMetricResource)(nil)
	_ resource.ResourceWithConfigure   = (*factMetricResource)(nil)
	_ resource.ResourceWithImportState = (*factMetricResource)(nil)
)

func newFactMetricResource() resource.Resource { return &factMetricResource{} }

type factMetricResource struct {
	client *client.Client
}

type factMetricResourceModel struct {
	ID                           types.String                       `tfsdk:"id"`
	Name                         types.String                       `tfsdk:"name"`
	Description                  types.String                       `tfsdk:"description"`
	Owner                        types.String                       `tfsdk:"owner"`
	Projects                     types.List                         `tfsdk:"projects"`
	Tags                         types.List                         `tfsdk:"tags"`
	Datasource                   types.String                       `tfsdk:"datasource"`
	MetricType                   types.String                       `tfsdk:"metric_type"`
	Numerator                    *factMetricColumnRefModel          `tfsdk:"numerator"`
	Denominator                  *factMetricDenominatorModel        `tfsdk:"denominator"`
	Inverse                      types.Bool                         `tfsdk:"inverse"`
	QuantileSettings             *factMetricQuantileSettingsModel   `tfsdk:"quantile_settings"`
	CappingSettings              *factMetricCappingSettingsModel    `tfsdk:"capping_settings"`
	WindowSettings               *factMetricWindowSettingsModel     `tfsdk:"window_settings"`
	PriorSettings                *factMetricPriorSettingsModel      `tfsdk:"prior_settings"`
	RegressionAdjustmentSettings *factMetricRegressionSettingsModel `tfsdk:"regression_adjustment_settings"`
	DisplayAsPercentage          types.Bool                         `tfsdk:"display_as_percentage"`
	MinPercentChange             types.Float64                      `tfsdk:"min_percent_change"`
	MaxPercentChange             types.Float64                      `tfsdk:"max_percent_change"`
	MinSampleSize                types.Float64                      `tfsdk:"min_sample_size"`
	TargetMDE                    types.Float64                      `tfsdk:"target_mde"`
	ManagedBy                    types.String                       `tfsdk:"managed_by"`
	MetricAutoSlices             types.List                         `tfsdk:"metric_auto_slices"`
	Archived                     types.Bool                         `tfsdk:"archived"`
	DateCreated                  types.String                       `tfsdk:"date_created"`
	DateUpdated                  types.String                       `tfsdk:"date_updated"`
}

// factMetricColumnRefModel is the numerator column reference (includes the
// numerator-only aggregate-filter attributes).
type factMetricColumnRefModel struct {
	FactTableID           types.String `tfsdk:"fact_table_id"`
	Column                types.String `tfsdk:"column"`
	Aggregation           types.String `tfsdk:"aggregation"`
	Filters               types.List   `tfsdk:"filters"`
	AggregateFilterColumn types.String `tfsdk:"aggregate_filter_column"`
	AggregateFilter       types.String `tfsdk:"aggregate_filter"`
}

// factMetricDenominatorModel is the denominator column reference. The API does
// not accept aggregate-filter fields on the denominator, so they are omitted.
type factMetricDenominatorModel struct {
	FactTableID types.String `tfsdk:"fact_table_id"`
	Column      types.String `tfsdk:"column"`
	Aggregation types.String `tfsdk:"aggregation"`
	Filters     types.List   `tfsdk:"filters"`
}

type factMetricCappingSettingsModel struct {
	Type        types.String  `tfsdk:"type"`
	Value       types.Float64 `tfsdk:"value"`
	IgnoreZeros types.Bool    `tfsdk:"ignore_zeros"`
}

type factMetricWindowSettingsModel struct {
	Type        types.String  `tfsdk:"type"`
	DelayValue  types.Float64 `tfsdk:"delay_value"`
	DelayUnit   types.String  `tfsdk:"delay_unit"`
	WindowValue types.Float64 `tfsdk:"window_value"`
	WindowUnit  types.String  `tfsdk:"window_unit"`
}

type factMetricPriorSettingsModel struct {
	Override types.Bool    `tfsdk:"override"`
	Proper   types.Bool    `tfsdk:"proper"`
	Mean     types.Float64 `tfsdk:"mean"`
	Stddev   types.Float64 `tfsdk:"stddev"`
}

type factMetricRegressionSettingsModel struct {
	Override types.Bool    `tfsdk:"override"`
	Enabled  types.Bool    `tfsdk:"enabled"`
	Days     types.Float64 `tfsdk:"days"`
}

type factMetricQuantileSettingsModel struct {
	Type                     types.String  `tfsdk:"type"`
	IgnoreZeros              types.Bool    `tfsdk:"ignore_zeros"`
	Quantile                 types.Float64 `tfsdk:"quantile"`
	QuantileEventCountColumn types.String  `tfsdk:"quantile_event_count_column"`
}

func (r *factMetricResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fact_metric"
}

func (r *factMetricResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *factMetricResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// factMetricColumnRefSchema builds the numerator/denominator column reference.
// The aggregate-filter attributes are numerator-only per the API, so they are
// only exposed when allowAggregateFilter is true (the denominator omits them).
func factMetricColumnRefSchema(required, allowAggregateFilter bool) schema.SingleNestedAttribute {
	attrs := map[string]schema.Attribute{
		"fact_table_id": schema.StringAttribute{Required: true, Description: "Fact table ID this metric column references."},
		"column": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Column name, or a special value such as `$$distinctUsers` or `$$count`. Must be empty for proportion metrics.",
		},
		"aggregation": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "User aggregation of the column: `sum`, `max`, `count distinct`, `hll merge`, or `kll merge`.",
		},
		"filters": schema.ListAttribute{
			Optional:    true,
			Computed:    true,
			ElementType: types.StringType,
			Description: "Fact table filter IDs applied to this column.",
		},
	}
	if allowAggregateFilter {
		attrs["aggregate_filter_column"] = schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Column used to filter users after aggregation (numerator only; retention/proportion metrics).",
		}
		attrs["aggregate_filter"] = schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Comparison applied after aggregation, e.g. `>= 1` (numerator only). Requires aggregate_filter_column.",
		}
	}
	return schema.SingleNestedAttribute{
		Required:    required,
		Optional:    !required,
		Description: "Reference to a fact table column, including aggregation and filters.",
		Attributes:  attrs,
	}
}

func (r *factMetricResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook fact metric built on top of a fact table. Supports " +
			"proportion, mean, ratio, quantile, retention, and dailyParticipation metric types.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique fact metric identifier.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name":        schema.StringAttribute{Required: true, Description: "Human-readable fact metric name."},
			"description": schema.StringAttribute{Optional: true, Computed: true, Description: "Description of the fact metric."},
			"owner":       schema.StringAttribute{Optional: true, Computed: true, Description: "Owner userId or email address."},
			"projects": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Associated project IDs.",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags associated with the fact metric.",
			},
			"datasource": schema.StringAttribute{
				Computed:    true,
				Description: "Data source ID (derived from the numerator's fact table).",
			},
			"metric_type": schema.StringAttribute{
				Required:    true,
				Description: "Metric type: `proportion`, `retention`, `mean`, `quantile`, `ratio`, or `dailyParticipation`.",
			},
			"numerator":   factMetricColumnRefSchema(true, true),
			"denominator": factMetricColumnRefSchema(false, false),
			"inverse": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Set true when a decrease is good (e.g. bounce rate).",
			},
			"quantile_settings": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Settings for quantile metrics (required when metric_type is `quantile`). Write-only: not refreshed from the server.",
				Attributes: map[string]schema.Attribute{
					"type":         schema.StringAttribute{Optional: true, Description: "Quantile over `event` values or `unit` aggregations."},
					"ignore_zeros": schema.BoolAttribute{Optional: true, Description: "Ignore zero values when calculating the quantile."},
					"quantile":     schema.Float64Attribute{Optional: true, Description: "Quantile value (0.001 to 0.999)."},
					"quantile_event_count_column": schema.StringAttribute{
						Optional:    true,
						Description: "Override for the per-row event count column (event quantile metrics with a `kll merge` numerator).",
					},
				},
			},
			"capping_settings": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Controls how outliers are handled. Write-only: not refreshed from the server.",
				Attributes: map[string]schema.Attribute{
					"type":         schema.StringAttribute{Optional: true, Description: "Capping type: `none`, `absolute`, or `percentile`."},
					"value":        schema.Float64Attribute{Optional: true, Description: "Absolute value or percentile (0.0-1.0) depending on type."},
					"ignore_zeros": schema.BoolAttribute{Optional: true, Description: "Ignore zeros when computing a percentile cap."},
				},
			},
			"window_settings": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Controls the conversion window for the metric. Write-only: not refreshed from the server.",
				Attributes: map[string]schema.Attribute{
					"type":         schema.StringAttribute{Optional: true, Description: "Window type: `none`, `conversion`, or `lookback`."},
					"delay_value":  schema.Float64Attribute{Optional: true, Description: "Delay after exposure before counting conversions."},
					"delay_unit":   schema.StringAttribute{Optional: true, Description: "Delay unit: `minutes`, `hours`, `days`, or `weeks`."},
					"window_value": schema.Float64Attribute{Optional: true, Description: "Length of the conversion window."},
					"window_unit":  schema.StringAttribute{Optional: true, Description: "Window unit: `minutes`, `hours`, `days`, or `weeks`."},
				},
			},
			"prior_settings": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Controls the bayesian prior for the metric. Write-only: not refreshed from the server.",
				Attributes: map[string]schema.Attribute{
					"override": schema.BoolAttribute{Optional: true, Description: "If false, organization default prior settings are used."},
					"proper":   schema.BoolAttribute{Optional: true, Description: "If true, use the configured mean and stddev; otherwise use an improper flat prior."},
					"mean":     schema.Float64Attribute{Optional: true, Description: "Prior mean of relative effects, in proportion terms."},
					"stddev":   schema.Float64Attribute{Optional: true, Description: "Prior standard deviation of relative effects (> 0)."},
				},
			},
			"regression_adjustment_settings": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Controls regression adjustment (CUPED) settings. Write-only: not refreshed from the server.",
				Attributes: map[string]schema.Attribute{
					"override": schema.BoolAttribute{Optional: true, Description: "If false, organization default settings are used."},
					"enabled":  schema.BoolAttribute{Optional: true, Description: "Whether regression adjustment is applied."},
					"days":     schema.Float64Attribute{Optional: true, Description: "Number of pre-exposure days used for the adjustment."},
				},
			},
			"display_as_percentage": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Display ratio/dailyParticipation means as a percentage.",
			},
			"min_percent_change": schema.Float64Attribute{Optional: true, Computed: true, Description: "Minimum percent change considered significant (proportion, e.g. 0.005 for 0.5%)."},
			"max_percent_change": schema.Float64Attribute{Optional: true, Computed: true, Description: "Maximum percent change considered significant (proportion, e.g. 0.5 for 50%)."},
			"min_sample_size":    schema.Float64Attribute{Optional: true, Computed: true, Description: "Minimum sample size before showing results."},
			"target_mde":         schema.Float64Attribute{Optional: true, Computed: true, Description: "Target minimum detectable effect (proportion, e.g. 0.1 for 10%)."},
			"managed_by": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Where the metric is managed from: empty, `api`, or `admin`.",
			},
			"metric_auto_slices": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Slice column names automatically included in metric analysis (enterprise).",
			},
			"archived":     schema.BoolAttribute{Optional: true, Computed: true, Description: "Whether the fact metric is archived."},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

// ---- build (model -> client) -------------------------------------------------

func factMetricColumnRefToClient(ctx context.Context, m *factMetricColumnRefModel, diags *diagAppender) *client.FactMetricColumnRef {
	if m == nil {
		return nil
	}
	return &client.FactMetricColumnRef{
		FactTableID:           m.FactTableID.ValueString(),
		Column:                m.Column.ValueString(),
		Aggregation:           m.Aggregation.ValueString(),
		Filters:               diags.strings(ctx, m.Filters),
		AggregateFilterColumn: m.AggregateFilterColumn.ValueString(),
		AggregateFilter:       m.AggregateFilter.ValueString(),
	}
}

func factMetricDenominatorToClient(ctx context.Context, m *factMetricDenominatorModel, diags *diagAppender) *client.FactMetricColumnRef {
	if m == nil {
		return nil
	}
	return &client.FactMetricColumnRef{
		FactTableID: m.FactTableID.ValueString(),
		Column:      m.Column.ValueString(),
		Aggregation: m.Aggregation.ValueString(),
		Filters:     diags.strings(ctx, m.Filters),
	}
}

func (m factMetricResourceModel) toInput(ctx context.Context, diags *diagAppender, isCreate bool) client.FactMetricInput {
	in := client.FactMetricInput{
		Name:                m.Name.ValueString(),
		Description:         optString(m.Description),
		Owner:               optString(m.Owner),
		Projects:            diags.strings(ctx, m.Projects),
		Tags:                diags.strings(ctx, m.Tags),
		MetricType:          m.MetricType.ValueString(),
		Numerator:           factMetricColumnRefToClient(ctx, m.Numerator, diags),
		Denominator:         factMetricDenominatorToClient(ctx, m.Denominator, diags),
		Inverse:             optBool(m.Inverse),
		DisplayAsPercentage: optBool(m.DisplayAsPercentage),
		MinPercentChange:    optFloat64(m.MinPercentChange),
		MaxPercentChange:    optFloat64(m.MaxPercentChange),
		MinSampleSize:       optFloat64(m.MinSampleSize),
		TargetMDE:           optFloat64(m.TargetMDE),
		ManagedBy:           optString(m.ManagedBy),
		MetricAutoSlices:    diags.strings(ctx, m.MetricAutoSlices),
	}
	// Archived is only accepted on update, not on create.
	if !isCreate {
		in.Archived = optBool(m.Archived)
	}
	if m.QuantileSettings != nil {
		in.QuantileSettings = &client.FactMetricQuantileSettings{
			Type:                     m.QuantileSettings.Type.ValueString(),
			IgnoreZeros:              m.QuantileSettings.IgnoreZeros.ValueBool(),
			Quantile:                 m.QuantileSettings.Quantile.ValueFloat64(),
			QuantileEventCountColumn: m.QuantileSettings.QuantileEventCountColumn.ValueString(),
		}
	}
	if m.CappingSettings != nil {
		in.CappingSettings = &client.FactMetricCappingSettings{
			Type:        m.CappingSettings.Type.ValueString(),
			Value:       optFloat64(m.CappingSettings.Value),
			IgnoreZeros: optBool(m.CappingSettings.IgnoreZeros),
		}
	}
	if m.WindowSettings != nil {
		in.WindowSettings = &client.FactMetricWindowSettings{
			Type:        m.WindowSettings.Type.ValueString(),
			DelayValue:  optFloat64(m.WindowSettings.DelayValue),
			DelayUnit:   m.WindowSettings.DelayUnit.ValueString(),
			WindowValue: optFloat64(m.WindowSettings.WindowValue),
			WindowUnit:  m.WindowSettings.WindowUnit.ValueString(),
		}
	}
	if m.PriorSettings != nil {
		in.PriorSettings = &client.FactMetricPriorSettings{
			Override: m.PriorSettings.Override.ValueBool(),
			Proper:   m.PriorSettings.Proper.ValueBool(),
			Mean:     m.PriorSettings.Mean.ValueFloat64(),
			Stddev:   m.PriorSettings.Stddev.ValueFloat64(),
		}
	}
	if m.RegressionAdjustmentSettings != nil {
		in.RegressionAdjustmentSettings = &client.FactMetricRegressionAdjustmentSettings{
			Override: m.RegressionAdjustmentSettings.Override.ValueBool(),
			Enabled:  optBool(m.RegressionAdjustmentSettings.Enabled),
			Days:     optFloat64(m.RegressionAdjustmentSettings.Days),
		}
	}
	return in
}

// ---- flatten (client -> model) -----------------------------------------------

func factMetricColumnRefToModel(ref *client.FactMetricColumnRef) *factMetricColumnRefModel {
	if ref == nil {
		return nil
	}
	return &factMetricColumnRefModel{
		FactTableID:           types.StringValue(ref.FactTableID),
		Column:                types.StringValue(ref.Column),
		Aggregation:           types.StringValue(ref.Aggregation),
		Filters:               sliceToStringList(ref.Filters),
		AggregateFilterColumn: types.StringValue(ref.AggregateFilterColumn),
		AggregateFilter:       types.StringValue(ref.AggregateFilter),
	}
}

func factMetricDenominatorToModel(ref *client.FactMetricColumnRef) *factMetricDenominatorModel {
	if ref == nil {
		return nil
	}
	return &factMetricDenominatorModel{
		FactTableID: types.StringValue(ref.FactTableID),
		Column:      types.StringValue(ref.Column),
		Aggregation: types.StringValue(ref.Aggregation),
		Filters:     sliceToStringList(ref.Filters),
	}
}

func (r *factMetricResource) apply(state *factMetricResourceModel, fm *client.FactMetric) {
	state.ID = types.StringValue(fm.ID)
	state.Name = types.StringValue(fm.Name)
	state.Description = types.StringValue(fm.Description)
	state.Owner = types.StringValue(fm.Owner)
	state.Projects = sliceToStringList(fm.Projects)
	state.Tags = sliceToStringList(fm.Tags)
	state.Datasource = types.StringValue(fm.Datasource)
	state.MetricType = types.StringValue(fm.MetricType)
	state.Numerator = factMetricColumnRefToModel(fm.Numerator)
	state.Denominator = factMetricDenominatorToModel(fm.Denominator)
	state.Inverse = types.BoolValue(fm.Inverse)
	state.DisplayAsPercentage = types.BoolValue(fm.DisplayAsPercentage)
	state.MinPercentChange = types.Float64Value(fm.MinPercentChange)
	state.MaxPercentChange = types.Float64Value(fm.MaxPercentChange)
	state.MinSampleSize = types.Float64Value(fm.MinSampleSize)
	state.TargetMDE = types.Float64Value(fm.TargetMDE)
	state.ManagedBy = types.StringValue(fm.ManagedBy)
	state.MetricAutoSlices = sliceToStringList(fm.MetricAutoSlices)
	state.Archived = types.BoolValue(fm.Archived)
	state.DateCreated = types.StringValue(fm.DateCreated)
	state.DateUpdated = types.StringValue(fm.DateUpdated)

	// quantile/capping/window/prior/regression settings are write-only: they are
	// preserved from configuration rather than read back, because the server
	// returns defaulted values that would otherwise create perpetual diffs.
}

func (r *factMetricResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan factMetricResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da, true)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateFactMetric(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create fact metric", err.Error())
		return
	}
	r.apply(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *factMetricResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state factMetricResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fm, err := r.client.GetFactMetric(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read fact metric", err.Error())
		return
	}
	r.apply(&state, fm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *factMetricResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan factMetricResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state factMetricResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da, false)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateFactMetric(ctx, state.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update fact metric", err.Error())
		return
	}
	r.apply(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *factMetricResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state factMetricResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteFactMetric(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete fact metric", err.Error())
	}
}
