package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

var (
	_ resource.Resource                = (*factTableFilterResource)(nil)
	_ resource.ResourceWithConfigure   = (*factTableFilterResource)(nil)
	_ resource.ResourceWithImportState = (*factTableFilterResource)(nil)
)

func newFactTableFilterResource() resource.Resource { return &factTableFilterResource{} }

type factTableFilterResource struct {
	client *client.Client
}

type factTableFilterResourceModel struct {
	ID          types.String `tfsdk:"id"`
	FactTableID types.String `tfsdk:"fact_table_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Value       types.String `tfsdk:"value"`
	ManagedBy   types.String `tfsdk:"managed_by"`
	DateCreated types.String `tfsdk:"date_created"`
	DateUpdated types.String `tfsdk:"date_updated"`
}

func (r *factTableFilterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fact_table_filter"
}

func (r *factTableFilterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *factTableFilterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook fact table filter: a named, reusable SQL expression " +
			"scoped to a single fact table.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique fact table filter identifier.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"fact_table_id": schema.StringAttribute{
				Required:      true,
				Description:   "ID of the fact table this filter belongs to. Changing this forces a new filter.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable filter name.",
			},
			"description": schema.StringAttribute{Optional: true, Computed: true, Description: "Description of the filter."},
			"value": schema.StringAttribute{
				Required:    true,
				Description: "SQL expression for this filter (e.g. `country = 'US'`).",
			},
			"managed_by": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Where the filter is managed from: empty or `api`. Set to `api` to disable editing in the GrowthBook UI (the fact table must also be `api`).",
			},
			"date_created": schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
			"date_updated": schema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
		},
	}
}

func (m factTableFilterResourceModel) toInput() client.FactTableFilterInput {
	return client.FactTableFilterInput{
		Name:        m.Name.ValueString(),
		Description: optString(m.Description),
		Value:       m.Value.ValueString(),
		ManagedBy:   optString(m.ManagedBy),
	}
}

func (r *factTableFilterResource) apply(state *factTableFilterResourceModel, factTableID string, f *client.FactTableFilter) {
	state.ID = types.StringValue(f.ID)
	state.FactTableID = types.StringValue(factTableID)
	state.Name = types.StringValue(f.Name)
	state.Description = types.StringValue(f.Description)
	state.Value = types.StringValue(f.Value)
	state.ManagedBy = types.StringValue(f.ManagedBy)
	state.DateCreated = types.StringValue(f.DateCreated)
	state.DateUpdated = types.StringValue(f.DateUpdated)
}

func (r *factTableFilterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan factTableFilterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	factTableID := plan.FactTableID.ValueString()
	created, err := r.client.CreateFactTableFilter(ctx, factTableID, plan.toInput())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create fact table filter", err.Error())
		return
	}
	r.apply(&plan, factTableID, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *factTableFilterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state factTableFilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	factTableID := state.FactTableID.ValueString()
	f, err := r.client.GetFactTableFilter(ctx, factTableID, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read fact table filter", err.Error())
		return
	}
	r.apply(&state, factTableID, f)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *factTableFilterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan factTableFilterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state factTableFilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	factTableID := state.FactTableID.ValueString()
	updated, err := r.client.UpdateFactTableFilter(ctx, factTableID, state.ID.ValueString(), plan.toInput())
	if err != nil {
		resp.Diagnostics.AddError("Unable to update fact table filter", err.Error())
		return
	}
	r.apply(&plan, factTableID, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *factTableFilterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state factTableFilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteFactTableFilter(ctx, state.FactTableID.ValueString(), state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete fact table filter", err.Error())
	}
}

// ImportState accepts a composite ID of the form "<factTableId>/<filterId>" and
// populates both the fact_table_id and id attributes.
func (r *factTableFilterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import ID in the format \"<fact_table_id>/<filter_id>\", got %q.", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("fact_table_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
