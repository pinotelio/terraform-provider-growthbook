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
	_ resource.Resource                = (*dimensionResource)(nil)
	_ resource.ResourceWithConfigure   = (*dimensionResource)(nil)
	_ resource.ResourceWithImportState = (*dimensionResource)(nil)
)

func newDimensionResource() resource.Resource { return &dimensionResource{} }

type dimensionResource struct {
	client *client.Client
}

type dimensionResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Owner          types.String `tfsdk:"owner"`
	Description    types.String `tfsdk:"description"`
	DatasourceID   types.String `tfsdk:"datasource_id"`
	IdentifierType types.String `tfsdk:"identifier_type"`
	Query          types.String `tfsdk:"query"`
	ManagedBy      types.String `tfsdk:"managed_by"`
	DateCreated    types.String `tfsdk:"date_created"`
	DateUpdated    types.String `tfsdk:"date_updated"`
}

func (r *dimensionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dimension"
}

func (r *dimensionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook user dimension. Dimensions slice experiment results by a " +
			"user attribute derived from a SQL query against a data source.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique dimension identifier (e.g. `dim_...`).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable dimension name.",
			},
			"owner": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "User ID or email of the dimension owner.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional description of the dimension.",
			},
			"datasource_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the data source this dimension queries.",
			},
			"identifier_type": schema.StringAttribute{
				Required:    true,
				Description: "Identifier type the dimension applies to (e.g. `user`, `anonymous`).",
			},
			"query": schema.StringAttribute{
				Required:    true,
				Description: "SQL query (or equivalent) that defines the dimension.",
			},
			"managed_by": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Where the dimension is managed from: empty (anywhere) or `api`.",
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

func (r *dimensionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (m dimensionResourceModel) toInput() client.DimensionInput {
	return client.DimensionInput{
		Name:           m.Name.ValueString(),
		Owner:          optString(m.Owner),
		Description:    optString(m.Description),
		DatasourceID:   m.DatasourceID.ValueString(),
		IdentifierType: m.IdentifierType.ValueString(),
		Query:          m.Query.ValueString(),
		ManagedBy:      optString(m.ManagedBy),
	}
}

func (r *dimensionResource) apply(state *dimensionResourceModel, d *client.Dimension) {
	state.ID = types.StringValue(d.ID)
	state.Name = types.StringValue(d.Name)
	state.Owner = types.StringValue(d.Owner)
	state.Description = types.StringValue(d.Description)
	state.DatasourceID = types.StringValue(d.DatasourceID)
	state.IdentifierType = types.StringValue(d.IdentifierType)
	state.Query = types.StringValue(d.Query)
	state.ManagedBy = types.StringValue(d.ManagedBy)
	state.DateCreated = types.StringValue(d.DateCreated)
	state.DateUpdated = types.StringValue(d.DateUpdated)
}

func (r *dimensionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dimensionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateDimension(ctx, plan.toInput())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create dimension", err.Error())
		return
	}
	r.apply(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dimensionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dimensionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	dim, err := r.client.GetDimension(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read dimension", err.Error())
		return
	}
	r.apply(&state, dim)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dimensionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dimensionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state dimensionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateDimension(ctx, state.ID.ValueString(), plan.toInput())
	if err != nil {
		resp.Diagnostics.AddError("Unable to update dimension", err.Error())
		return
	}
	r.apply(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dimensionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dimensionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteDimension(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete dimension", err.Error())
	}
}

func (r *dimensionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
