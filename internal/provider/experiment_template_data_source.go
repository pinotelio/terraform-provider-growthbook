package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ datasource.DataSource              = (*experimentTemplateDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*experimentTemplateDataSource)(nil)
)

func newExperimentTemplateDataSource() datasource.DataSource { return &experimentTemplateDataSource{} }

type experimentTemplateDataSource struct {
	client *client.Client
}

func (d *experimentTemplateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_experiment_template"
}

func (d *experimentTemplateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing GrowthBook experiment template by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Unique experiment template identifier.",
			},
			"project":     schema.StringAttribute{Computed: true, Description: "Project ID."},
			"owner":       schema.StringAttribute{Computed: true, Description: "Owner user ID."},
			"owner_email": schema.StringAttribute{Computed: true, Description: "Owner email, when resolvable."},
			"template_metadata": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Template name and description.",
				Attributes: map[string]schema.Attribute{
					"name":        schema.StringAttribute{Computed: true, Description: "Template name."},
					"description": schema.StringAttribute{Computed: true, Description: "Template description."},
				},
			},
			"type":                     schema.StringAttribute{Computed: true, Description: "Template type."},
			"hypothesis":               schema.StringAttribute{Computed: true, Description: "Experiment hypothesis."},
			"description":              schema.StringAttribute{Computed: true, Description: "Experiment description."},
			"tags":                     schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Tags."},
			"datasource":               schema.StringAttribute{Computed: true, Description: "Datasource ID."},
			"exposure_query_id":        schema.StringAttribute{Computed: true, Description: "Exposure query ID."},
			"hash_attribute":           schema.StringAttribute{Computed: true, Description: "Bucketing attribute."},
			"fallback_attribute":       schema.StringAttribute{Computed: true, Description: "Fallback bucketing attribute."},
			"disable_sticky_bucketing": schema.BoolAttribute{Computed: true, Description: "Whether sticky bucketing is disabled."},
			"goal_metrics":             schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Goal metric IDs."},
			"secondary_metrics":        schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Secondary metric IDs."},
			"guardrail_metrics":        schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Guardrail metric IDs."},
			"activation_metric":        schema.StringAttribute{Computed: true, Description: "Activation metric ID."},
			"stats_engine":             schema.StringAttribute{Computed: true, Description: "Statistics engine."},
			"segment":                  schema.StringAttribute{Computed: true, Description: "Segment ID."},
			"skip_partial_data":        schema.BoolAttribute{Computed: true, Description: "Whether partial data is skipped."},
			"targeting": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Default targeting configuration.",
				Attributes: map[string]schema.Attribute{
					"coverage":  schema.Float64Attribute{Computed: true, Description: "Traffic coverage (0-1)."},
					"condition": schema.StringAttribute{Computed: true, Description: "Targeting condition (JSON string)."},
					"saved_groups": schema.ListNestedAttribute{
						Computed:    true,
						Description: "Saved group targeting clauses.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"match": schema.StringAttribute{Computed: true, Description: "Match type."},
								"ids":   schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Saved group IDs."},
							},
						},
					},
					"prerequisites": schema.ListNestedAttribute{
						Computed:    true,
						Description: "Feature prerequisites.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id":        schema.StringAttribute{Computed: true, Description: "Prerequisite feature ID."},
								"condition": schema.StringAttribute{Computed: true, Description: "Prerequisite condition (JSON string)."},
							},
						},
					},
				},
			},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

func (d *experimentTemplateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *experimentTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data experimentTemplateResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	t, err := d.client.GetExperimentTemplate(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read experiment template", err.Error())
		return
	}
	r := &experimentTemplateResource{}
	r.apply(&data, t)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
