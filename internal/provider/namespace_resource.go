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
	_ resource.Resource                = (*namespaceResource)(nil)
	_ resource.ResourceWithConfigure   = (*namespaceResource)(nil)
	_ resource.ResourceWithImportState = (*namespaceResource)(nil)
)

func newNamespaceResource() resource.Resource { return &namespaceResource{} }

type namespaceResource struct {
	client *client.Client
}

type namespaceResourceModel struct {
	ID            types.String `tfsdk:"id"`
	DisplayName   types.String `tfsdk:"display_name"`
	Description   types.String `tfsdk:"description"`
	Status        types.String `tfsdk:"status"`
	Format        types.String `tfsdk:"format"`
	HashAttribute types.String `tfsdk:"hash_attribute"`
	Seed          types.String `tfsdk:"seed"`
}

func (r *namespaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespace"
}

func (r *namespaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook namespace. Namespaces partition experiment traffic so that " +
			"mutually exclusive experiments never overlap.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique namespace identifier (e.g. `ns-...`).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"display_name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable display name. Must be unique within the organization.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the namespace.",
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Namespace status: `active` or `inactive`.",
			},
			"format": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "Namespace format: `legacy` or `multiRange`. Defaults to `multiRange`. " +
					"Immutable; changing this forces a new namespace.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hash_attribute": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "The user attribute used for bucket hashing. Required when `format` " +
					"is `multiRange`.",
			},
			"seed": schema.StringAttribute{
				Computed:    true,
				Description: "The seed used for bucket hashing. Managed by GrowthBook.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *namespaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *namespaceResource) apply(state *namespaceResourceModel, n *client.Namespace) {
	state.ID = types.StringValue(n.ID)
	state.DisplayName = types.StringValue(n.DisplayName)
	state.Description = types.StringValue(n.Description)
	state.Status = types.StringValue(n.Status)
	state.Format = types.StringValue(n.Format)
	if n.HashAttribute == "" {
		state.HashAttribute = types.StringNull()
	} else {
		state.HashAttribute = types.StringValue(n.HashAttribute)
	}
	if n.Seed == "" {
		state.Seed = types.StringNull()
	} else {
		state.Seed = types.StringValue(n.Seed)
	}
}

func (r *namespaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan namespaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.NamespaceCreateInput{
		DisplayName:   plan.DisplayName.ValueString(),
		Description:   optString(plan.Description),
		Status:        optString(plan.Status),
		Format:        optString(plan.Format),
		HashAttribute: optString(plan.HashAttribute),
	}

	created, err := r.client.CreateNamespace(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create namespace", err.Error())
		return
	}
	r.apply(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *namespaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state namespaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ns, err := r.client.GetNamespace(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read namespace", err.Error())
		return
	}
	r.apply(&state, ns)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *namespaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan namespaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state namespaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.NamespaceUpdateInput{
		DisplayName:   optString(plan.DisplayName),
		Description:   optString(plan.Description),
		Status:        optString(plan.Status),
		HashAttribute: optString(plan.HashAttribute),
	}

	updated, err := r.client.UpdateNamespace(ctx, state.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update namespace", err.Error())
		return
	}
	r.apply(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *namespaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state namespaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteNamespace(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete namespace", err.Error())
	}
}

func (r *namespaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
