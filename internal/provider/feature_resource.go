package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ resource.Resource                = (*featureResource)(nil)
	_ resource.ResourceWithConfigure   = (*featureResource)(nil)
	_ resource.ResourceWithImportState = (*featureResource)(nil)
)

func newFeatureResource() resource.Resource { return &featureResource{} }

type featureResource struct {
	client *client.Client
}

type featureResourceModel struct {
	ID            types.String               `tfsdk:"id"`
	Description   types.String               `tfsdk:"description"`
	Owner         types.String               `tfsdk:"owner"`
	Project       types.String               `tfsdk:"project"`
	Archived      types.Bool                 `tfsdk:"archived"`
	ValueType     types.String               `tfsdk:"value_type"`
	DefaultValue  types.String               `tfsdk:"default_value"`
	Tags          types.List                 `tfsdk:"tags"`
	Prerequisites types.List                 `tfsdk:"prerequisites"`
	Environments  map[string]featureEnvModel `tfsdk:"environments"`
	DateCreated   types.String               `tfsdk:"date_created"`
	DateUpdated   types.String               `tfsdk:"date_updated"`
}

type featureEnvModel struct {
	Enabled types.Bool         `tfsdk:"enabled"`
	Rules   []featureRuleModel `tfsdk:"rules"`
}

type featureRuleModel struct {
	Type                   types.String                  `tfsdk:"type"`
	Description            types.String                  `tfsdk:"description"`
	Enabled                types.Bool                    `tfsdk:"enabled"`
	Condition              types.String                  `tfsdk:"condition"`
	Value                  types.String                  `tfsdk:"value"`
	Coverage               types.Float64                 `tfsdk:"coverage"`
	HashAttribute          types.String                  `tfsdk:"hash_attribute"`
	Seed                   types.String                  `tfsdk:"seed"`
	HashVersion            types.Int64                   `tfsdk:"hash_version"`
	TrackingKey            types.String                  `tfsdk:"tracking_key"`
	FallbackAttribute      types.String                  `tfsdk:"fallback_attribute"`
	DisableStickyBucketing types.Bool                    `tfsdk:"disable_sticky_bucketing"`
	BucketVersion          types.Int64                   `tfsdk:"bucket_version"`
	MinBucketVersion       types.Int64                   `tfsdk:"min_bucket_version"`
	ExperimentID           types.String                  `tfsdk:"experiment_id"`
	Sparse                 types.Bool                    `tfsdk:"sparse"`
	SavedGroupTargeting    []savedGroupTargetingModel    `tfsdk:"saved_group_targeting"`
	Prerequisites          []rulePrereqModel             `tfsdk:"prerequisites"`
	ScheduleRules          []scheduleRuleModel           `tfsdk:"schedule_rules"`
	Namespace              *ruleNamespaceModel           `tfsdk:"namespace"`
	Values                 []experimentValueModel        `tfsdk:"values"`
	Variations             []experimentRefVariationModel `tfsdk:"variations"`
}

type savedGroupTargetingModel struct {
	MatchType   types.String `tfsdk:"match_type"`
	SavedGroups types.List   `tfsdk:"saved_groups"`
}

type rulePrereqModel struct {
	ID        types.String `tfsdk:"id"`
	Condition types.String `tfsdk:"condition"`
}

type scheduleRuleModel struct {
	Enabled   types.Bool   `tfsdk:"enabled"`
	Timestamp types.String `tfsdk:"timestamp"`
}

type ruleNamespaceModel struct {
	Enabled  types.Bool    `tfsdk:"enabled"`
	Name     types.String  `tfsdk:"name"`
	RangeMin types.Float64 `tfsdk:"range_min"`
	RangeMax types.Float64 `tfsdk:"range_max"`
}

type experimentValueModel struct {
	Value  types.String  `tfsdk:"value"`
	Weight types.Float64 `tfsdk:"weight"`
	Name   types.String  `tfsdk:"name"`
}

type experimentRefVariationModel struct {
	Value       types.String `tfsdk:"value"`
	VariationID types.String `tfsdk:"variation_id"`
}

func (r *featureResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_feature"
}

func (r *featureResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *featureResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *featureResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook feature flag, including its value type, default value, " +
			"and per-environment enablement and targeting rules.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Feature key. May contain letters, numbers, hyphens and underscores. Changing this forces a new feature.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{Optional: true, Computed: true, Description: "Description of the feature."},
			"owner":       schema.StringAttribute{Optional: true, Computed: true, Description: "Owner userId or email address."},
			"project":     schema.StringAttribute{Optional: true, Computed: true, Description: "Associated project ID."},
			"archived":    schema.BoolAttribute{Optional: true, Computed: true, Description: "Whether the feature is archived."},
			"value_type": schema.StringAttribute{
				Required:    true,
				Description: "Data type of the feature value: `boolean`, `string`, `number`, or `json`.",
			},
			"default_value": schema.StringAttribute{
				Required:    true,
				Description: "Default value returned when the feature is enabled (encoded as a string).",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags associated with the feature.",
			},
			"prerequisites": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Feature IDs that must evaluate to `true` for this feature (read-only).",
			},
			"environments": schema.MapNestedAttribute{
				Optional:     true,
				Description:  "Per-environment configuration, keyed by environment ID.",
				NestedObject: featureEnvironmentSchema(),
			},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

func featureEnvironmentSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the feature is enabled in this environment.",
			},
			"rules": schema.ListNestedAttribute{
				Optional:     true,
				Description:  "Ordered list of targeting rules evaluated in this environment.",
				NestedObject: featureRuleSchema(),
			},
		},
	}
}

func featureRuleSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Rule type: `force`, `rollout`, `experiment`, or `experiment-ref`.",
			},
			"description": schema.StringAttribute{Optional: true, Description: "Human-readable rule description."},
			"enabled":     schema.BoolAttribute{Optional: true, Description: "Whether the rule is enabled."},
			"condition":   schema.StringAttribute{Optional: true, Description: "MongoDB-style targeting condition (JSON string)."},
			"value":       schema.StringAttribute{Optional: true, Description: "Value to serve (force/rollout rules)."},
			"coverage":    schema.Float64Attribute{Optional: true, Description: "Fraction of traffic included (rollout/experiment rules)."},
			"hash_attribute": schema.StringAttribute{
				Optional:    true,
				Description: "Attribute used for bucketing (rollout/experiment rules).",
			},
			"seed":         schema.StringAttribute{Optional: true, Description: "Hash seed for bucketing (rollout rules)."},
			"hash_version": schema.Int64Attribute{Optional: true, Description: "Hash algorithm version (1 or 2)."},
			"tracking_key": schema.StringAttribute{Optional: true, Description: "Experiment tracking key (experiment rules)."},
			"fallback_attribute": schema.StringAttribute{
				Optional:    true,
				Description: "Fallback bucketing attribute for sticky bucketing (experiment rules).",
			},
			"disable_sticky_bucketing": schema.BoolAttribute{Optional: true, Description: "Disable sticky bucketing (experiment rules)."},
			"bucket_version":           schema.Int64Attribute{Optional: true, Description: "Sticky bucket version (experiment rules)."},
			"min_bucket_version":       schema.Int64Attribute{Optional: true, Description: "Minimum sticky bucket version (experiment rules)."},
			"experiment_id":            schema.StringAttribute{Optional: true, Description: "Linked experiment ID (experiment-ref rules)."},
			"sparse": schema.BoolAttribute{
				Optional:    true,
				Description: "JSON features only: merge the rule value onto the default instead of replacing it.",
			},
			"saved_group_targeting": schema.ListNestedAttribute{
				Optional:     true,
				Description:  "Saved group targeting clauses.",
				NestedObject: savedGroupTargetingSchema(),
			},
			"prerequisites": schema.ListNestedAttribute{
				Optional:     true,
				Description:  "Per-rule feature prerequisites.",
				NestedObject: rulePrerequisiteSchema(),
			},
			"schedule_rules": schema.ListNestedAttribute{
				Optional:     true,
				Description:  "Time-based on/off schedule for the rule.",
				NestedObject: scheduleRuleSchema(),
			},
			"namespace": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Namespace partition for experiment rules.",
				Attributes: map[string]schema.Attribute{
					"enabled":   schema.BoolAttribute{Required: true, Description: "Whether namespace targeting is enabled."},
					"name":      schema.StringAttribute{Required: true, Description: "Namespace name."},
					"range_min": schema.Float64Attribute{Required: true, Description: "Lower bound of the namespace range (0-1)."},
					"range_max": schema.Float64Attribute{Required: true, Description: "Upper bound of the namespace range (0-1)."},
				},
			},
			"values": schema.ListNestedAttribute{
				Optional:     true,
				Description:  "Weighted variation values (experiment rules).",
				NestedObject: experimentValueSchema(),
			},
			"variations": schema.ListNestedAttribute{
				Optional:     true,
				Description:  "Variation-to-value mappings (experiment-ref rules).",
				NestedObject: experimentRefVariationSchema(),
			},
		},
	}
}

func savedGroupTargetingSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"match_type": schema.StringAttribute{Required: true, Description: "Match type: `all`, `any`, or `none`."},
			"saved_groups": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Saved group IDs.",
			},
		},
	}
}

func rulePrerequisiteSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id":        schema.StringAttribute{Required: true, Description: "Prerequisite feature ID."},
			"condition": schema.StringAttribute{Required: true, Description: "Condition the prerequisite feature must satisfy (JSON string)."},
		},
	}
}

func scheduleRuleSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"enabled":   schema.BoolAttribute{Required: true, Description: "Target enabled state at the timestamp."},
			"timestamp": schema.StringAttribute{Optional: true, Description: "ISO timestamp when the rule changes state. Null means immediately."},
		},
	}
}

func experimentValueSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"value":  schema.StringAttribute{Required: true, Description: "Variation value."},
			"weight": schema.Float64Attribute{Required: true, Description: "Traffic weight (0-1)."},
			"name":   schema.StringAttribute{Optional: true, Description: "Optional variation name."},
		},
	}
}

func experimentRefVariationSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"value":        schema.StringAttribute{Required: true, Description: "Feature value for the variation."},
			"variation_id": schema.StringAttribute{Required: true, Description: "Experiment variation ID."},
		},
	}
}
