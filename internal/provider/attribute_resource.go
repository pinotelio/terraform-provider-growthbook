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
	_ resource.Resource                = (*attributeResource)(nil)
	_ resource.ResourceWithConfigure   = (*attributeResource)(nil)
	_ resource.ResourceWithImportState = (*attributeResource)(nil)
)

func newAttributeResource() resource.Resource { return &attributeResource{} }

type attributeResource struct {
	client *client.Client
}

type attributeResourceModel struct {
	Property      types.String `tfsdk:"property"`
	Datatype      types.String `tfsdk:"datatype"`
	Description   types.String `tfsdk:"description"`
	HashAttribute types.Bool   `tfsdk:"hash_attribute"`
	Archived      types.Bool   `tfsdk:"archived"`
	Enum          types.String `tfsdk:"enum"`
	Format        types.String `tfsdk:"format"`
	Projects      types.List   `tfsdk:"projects"`
	Tags          types.List   `tfsdk:"tags"`
}

func (r *attributeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_attribute"
}

func (r *attributeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook targeting attribute. Attributes describe the properties of a " +
			"user (e.g. `id`, `country`) and are used to target feature flag rules and experiments.",
		Attributes: map[string]schema.Attribute{
			"property": schema.StringAttribute{
				Required: true,
				Description: "The attribute property name. This is the attribute's unique identifier. " +
					"Changing this forces a new attribute.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"datatype": schema.StringAttribute{
				Required: true,
				Description: "The attribute datatype. One of `boolean`, `string`, `number`, " +
					"`secureString`, `enum`, `string[]`, `number[]`, `secureString[]`.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the attribute.",
			},
			"hash_attribute": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether this attribute should be hashed when used for bucketing.",
			},
			"archived": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the attribute is archived.",
			},
			"enum": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Comma-separated list of allowed values when `datatype` is `enum`.",
			},
			"format": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The attribute's format. One of `` (none), `version`, `date`, `isoCountryCode`.",
			},
			"projects": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Project IDs this attribute is scoped to. Empty means all projects.",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Tags applied to the attribute.",
			},
		},
	}
}

func (r *attributeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *attributeResource) apply(state *attributeResourceModel, a *client.Attribute) {
	state.Property = types.StringValue(a.Property)
	state.Datatype = types.StringValue(a.Datatype)
	state.Description = types.StringValue(a.Description)
	state.HashAttribute = types.BoolValue(a.HashAttribute)
	state.Archived = types.BoolValue(a.Archived)
	state.Enum = types.StringValue(a.Enum)
	state.Format = types.StringValue(a.Format)
	state.Projects = sliceToStringList(a.Projects)
	state.Tags = sliceToStringList(a.Tags)
}

func (r *attributeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan attributeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.AttributeCreateInput{
		Property:      plan.Property.ValueString(),
		Datatype:      plan.Datatype.ValueString(),
		Description:   optString(plan.Description),
		Archived:      optBool(plan.Archived),
		HashAttribute: optBool(plan.HashAttribute),
		Enum:          optString(plan.Enum),
		Format:        optString(plan.Format),
		Projects:      stringListToSlice(ctx, plan.Projects, &resp.Diagnostics),
		Tags:          stringListToSlice(ctx, plan.Tags, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateAttribute(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create attribute", err.Error())
		return
	}
	r.apply(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *attributeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state attributeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attr, err := r.client.GetAttribute(ctx, state.Property.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read attribute", err.Error())
		return
	}
	r.apply(&state, attr)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *attributeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan attributeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.AttributeUpdateInput{
		Datatype:      optString(plan.Datatype),
		Description:   optString(plan.Description),
		Archived:      optBool(plan.Archived),
		HashAttribute: optBool(plan.HashAttribute),
		Enum:          optString(plan.Enum),
		Format:        optString(plan.Format),
		Projects:      stringListToSlice(ctx, plan.Projects, &resp.Diagnostics),
		Tags:          stringListToSlice(ctx, plan.Tags, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateAttribute(ctx, plan.Property.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update attribute", err.Error())
		return
	}
	r.apply(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *attributeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state attributeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAttribute(ctx, state.Property.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete attribute", err.Error())
	}
}

func (r *attributeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("property"), req, resp)
}
