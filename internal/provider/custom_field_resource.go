package provider

import (
	"context"
	"encoding/json"
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
	_ resource.Resource                = (*customFieldResource)(nil)
	_ resource.ResourceWithConfigure   = (*customFieldResource)(nil)
	_ resource.ResourceWithImportState = (*customFieldResource)(nil)
)

func newCustomFieldResource() resource.Resource { return &customFieldResource{} }

type customFieldResource struct {
	client *client.Client
}

type customFieldResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Placeholder  types.String `tfsdk:"placeholder"`
	DefaultValue types.String `tfsdk:"default_value"`
	Type         types.String `tfsdk:"type"`
	Values       types.String `tfsdk:"values"`
	Required     types.Bool   `tfsdk:"required"`
	Creator      types.String `tfsdk:"creator"`
	Projects     types.List   `tfsdk:"projects"`
	Sections     types.List   `tfsdk:"sections"`
	Active       types.Bool   `tfsdk:"active"`
	DateCreated  types.String `tfsdk:"date_created"`
	DateUpdated  types.String `tfsdk:"date_updated"`
}

func (r *customFieldResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_field"
}

func (r *customFieldResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A GrowthBook custom field. Custom fields add structured metadata to " +
			"features and experiments.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:      true,
				Description:   "Unique key for the custom field. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Display name of the custom field.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional description of the custom field.",
			},
			"placeholder": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Placeholder text shown in the input.",
			},
			"default_value": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "JSON-encoded default value. Encodes whichever shape the field type " +
					"accepts (e.g. `\"text\"`, `42`, `true`, or `[\"a\",\"b\"]`).",
			},
			"type": schema.StringAttribute{
				Required: true,
				Description: "Value type of the custom field: `text`, `textarea`, `markdown`, " +
					"`enum`, `multiselect`, `url`, `number`, `boolean`, `date`, or `datetime`. " +
					"Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"values": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Comma-separated list of allowed values for `enum`/`multiselect` types.",
			},
			"required": schema.BoolAttribute{
				Required:    true,
				Description: "Whether a value must be provided for this field.",
			},
			"creator": schema.StringAttribute{
				Computed:    true,
				Description: "User ID of the field creator.",
			},
			"projects": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Project IDs this field applies to. Empty means all projects.",
			},
			"sections": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Object types this field applies to: `feature` and/or `experiment`.",
			},
			"active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the custom field is active.",
			},
			"date_created": schema.StringAttribute{
				Computed:      true,
				Description:   "Creation timestamp (RFC3339).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"date_updated": schema.StringAttribute{
				Computed:    true,
				Description: "Last update timestamp (RFC3339).",
			},
		},
	}
}

func (r *customFieldResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (m customFieldResourceModel) toCreateInput(ctx context.Context, diags *diagAppender) client.CustomFieldCreateInput {
	return client.CustomFieldCreateInput{
		ID:           m.ID.ValueString(),
		Name:         m.Name.ValueString(),
		Description:  optString(m.Description),
		Placeholder:  optString(m.Placeholder),
		DefaultValue: customFieldDefaultValueToRaw(m.DefaultValue),
		Type:         m.Type.ValueString(),
		Values:       optString(m.Values),
		Required:     m.Required.ValueBool(),
		Projects:     diags.strings(ctx, m.Projects),
		Sections:     diags.strings(ctx, m.Sections),
	}
}

func (m customFieldResourceModel) toUpdateInput(ctx context.Context, diags *diagAppender) client.CustomFieldUpdateInput {
	return client.CustomFieldUpdateInput{
		Name:         m.Name.ValueString(),
		Description:  optString(m.Description),
		Placeholder:  optString(m.Placeholder),
		DefaultValue: customFieldDefaultValueToRaw(m.DefaultValue),
		Values:       optString(m.Values),
		Required:     optBool(m.Required),
		Projects:     diags.strings(ctx, m.Projects),
		Sections:     diags.strings(ctx, m.Sections),
		Active:       optBool(m.Active),
	}
}

func (r *customFieldResource) apply(state *customFieldResourceModel, f *client.CustomField) {
	state.ID = types.StringValue(f.ID)
	state.Name = types.StringValue(f.Name)
	state.Description = types.StringValue(f.Description)
	state.Placeholder = types.StringValue(f.Placeholder)
	state.DefaultValue = customFieldDefaultValueToString(f.DefaultValue)
	state.Type = types.StringValue(f.Type)
	state.Values = types.StringValue(f.Values)
	state.Required = types.BoolValue(f.Required)
	state.Creator = types.StringValue(f.Creator)
	state.Projects = sliceToStringList(f.Projects)
	state.Sections = sliceToStringList(f.Sections)
	state.Active = types.BoolValue(f.Active)
	state.DateCreated = types.StringValue(f.DateCreated)
	state.DateUpdated = types.StringValue(f.DateUpdated)
}

// customFieldDefaultValueToRaw converts the Terraform JSON string into raw JSON
// for the request body. Null/unknown/empty yields nil so the field is omitted.
func customFieldDefaultValueToRaw(v types.String) json.RawMessage {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	if s == "" {
		return nil
	}
	return json.RawMessage(s)
}

// customFieldDefaultValueToString converts raw JSON from the API into a
// Terraform string, using null when there is no payload.
func customFieldDefaultValueToString(raw json.RawMessage) types.String {
	if len(raw) == 0 {
		return types.StringNull()
	}
	return types.StringValue(string(raw))
}

func (r *customFieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customFieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toCreateInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateCustomField(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create custom field", err.Error())
		return
	}
	r.apply(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *customFieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customFieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	field, err := r.client.GetCustomField(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read custom field", err.Error())
		return
	}
	r.apply(&state, field)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *customFieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customFieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state customFieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toUpdateInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateCustomField(ctx, state.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update custom field", err.Error())
		return
	}
	r.apply(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *customFieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customFieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteCustomField(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete custom field", err.Error())
	}
}

func (r *customFieldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
