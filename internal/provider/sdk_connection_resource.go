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
	_ resource.Resource                = (*sdkConnectionResource)(nil)
	_ resource.ResourceWithConfigure   = (*sdkConnectionResource)(nil)
	_ resource.ResourceWithImportState = (*sdkConnectionResource)(nil)
)

func newSDKConnectionResource() resource.Resource { return &sdkConnectionResource{} }

type sdkConnectionResource struct {
	client *client.Client
}

type sdkConnectionResourceModel struct {
	ID                            types.String `tfsdk:"id"`
	Name                          types.String `tfsdk:"name"`
	Language                      types.String `tfsdk:"language"`
	Languages                     types.List   `tfsdk:"languages"`
	SDKVersion                    types.String `tfsdk:"sdk_version"`
	Environment                   types.String `tfsdk:"environment"`
	Projects                      types.List   `tfsdk:"projects"`
	EncryptPayload                types.Bool   `tfsdk:"encrypt_payload"`
	IncludeVisualExperiments      types.Bool   `tfsdk:"include_visual_experiments"`
	IncludeDraftExperiments       types.Bool   `tfsdk:"include_draft_experiments"`
	IncludeDraftExperimentRefs    types.Bool   `tfsdk:"include_draft_experiment_refs"`
	IncludeExperimentNames        types.Bool   `tfsdk:"include_experiment_names"`
	IncludeRedirectExperiments    types.Bool   `tfsdk:"include_redirect_experiments"`
	IncludeRuleIDs                types.Bool   `tfsdk:"include_rule_ids"`
	IncludeProjectIDInMetadata    types.Bool   `tfsdk:"include_project_id_in_metadata"`
	IncludeCustomFieldsInMetadata types.Bool   `tfsdk:"include_custom_fields_in_metadata"`
	AllowedCustomFieldsInMetadata types.List   `tfsdk:"allowed_custom_fields_in_metadata"`
	IncludeTagsInMetadata         types.Bool   `tfsdk:"include_tags_in_metadata"`
	ProxyEnabled                  types.Bool   `tfsdk:"proxy_enabled"`
	ProxyHost                     types.String `tfsdk:"proxy_host"`
	HashSecureAttributes          types.Bool   `tfsdk:"hash_secure_attributes"`
	RemoteEvalEnabled             types.Bool   `tfsdk:"remote_eval_enabled"`
	SavedGroupReferencesEnabled   types.Bool   `tfsdk:"saved_group_references_enabled"`

	// Computed-only output fields.
	Key             types.String `tfsdk:"key"`
	EncryptionKey   types.String `tfsdk:"encryption_key"`
	ProxySigningKey types.String `tfsdk:"proxy_signing_key"`
	SSEEnabled      types.Bool   `tfsdk:"sse_enabled"`
	Organization    types.String `tfsdk:"organization"`
}

func optionalComputedBool() schema.BoolAttribute {
	return schema.BoolAttribute{Optional: true, Computed: true}
}

func (r *sdkConnectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sdk_connection"
}

func (r *sdkConnectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	bools := map[string]string{
		"encrypt_payload":                   "Encrypt the SDK payload.",
		"include_visual_experiments":        "Include visual experiments in the payload.",
		"include_draft_experiments":         "Include draft experiments in the payload.",
		"include_draft_experiment_refs":     "Include experiment-ref rules linked to draft experiments.",
		"include_experiment_names":          "Include experiment and variation names in the payload.",
		"include_redirect_experiments":      "Include URL redirect experiments in the payload.",
		"include_rule_ids":                  "Include feature rule IDs in the payload.",
		"include_project_id_in_metadata":    "Include the project ID in payload metadata.",
		"include_custom_fields_in_metadata": "Include custom fields in payload metadata.",
		"include_tags_in_metadata":          "Include tags in payload metadata.",
		"proxy_enabled":                     "Enable the GrowthBook proxy for this connection.",
		"hash_secure_attributes":            "Hash secure attributes before sending to the SDK.",
		"remote_eval_enabled":               "Enable remote evaluation.",
		"saved_group_references_enabled":    "Enable saved group references in the payload.",
	}

	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:      true,
			Description:   "Unique SDK connection identifier.",
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": schema.StringAttribute{
			Required:    true,
			Description: "Name of the SDK connection.",
		},
		"language": schema.StringAttribute{
			Required:    true,
			Description: "Primary SDK language (e.g. `javascript`, `react`, `go`, `python`).",
		},
		"languages": schema.ListAttribute{
			Computed:    true,
			ElementType: types.StringType,
			Description: "All languages associated with the connection, as returned by the API.",
		},
		"sdk_version": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Pinned SDK version for payload compatibility.",
		},
		"environment": schema.StringAttribute{
			Required:    true,
			Description: "Environment this connection serves.",
		},
		"projects": schema.ListAttribute{
			Optional:    true,
			ElementType: types.StringType,
			Description: "Project IDs this connection is scoped to. Empty means all projects.",
		},
		"allowed_custom_fields_in_metadata": schema.ListAttribute{
			Optional:    true,
			ElementType: types.StringType,
			Description: "Custom field keys allowed in payload metadata.",
		},
		"proxy_host": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Proxy host URL when the proxy is enabled.",
		},

		// Computed-only.
		"key":               schema.StringAttribute{Computed: true, Description: "Client key used by SDKs to fetch features."},
		"encryption_key":    schema.StringAttribute{Computed: true, Sensitive: true, Description: "Payload encryption key (when encryption is enabled)."},
		"proxy_signing_key": schema.StringAttribute{Computed: true, Sensitive: true, Description: "Proxy signing key."},
		"sse_enabled":       schema.BoolAttribute{Computed: true, Description: "Whether server-sent events streaming is enabled."},
		"organization":      schema.StringAttribute{Computed: true, Description: "Owning organization ID."},
	}
	for name, desc := range bools {
		a := optionalComputedBool()
		a.Description = desc
		attrs[name] = a
	}

	resp.Schema = schema.Schema{
		Description: "A GrowthBook SDK connection. SDK connections expose feature and " +
			"experiment definitions to client/server SDKs for a given environment.",
		Attributes: attrs,
	}
}

func (r *sdkConnectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (m sdkConnectionResourceModel) toInput(ctx context.Context, diags *diagAppender) client.SDKConnectionInput {
	return client.SDKConnectionInput{
		Name:                          m.Name.ValueString(),
		Language:                      m.Language.ValueString(),
		SDKVersion:                    optString(m.SDKVersion),
		Environment:                   m.Environment.ValueString(),
		Projects:                      diags.strings(ctx, m.Projects),
		EncryptPayload:                optBool(m.EncryptPayload),
		IncludeVisualExperiments:      optBool(m.IncludeVisualExperiments),
		IncludeDraftExperiments:       optBool(m.IncludeDraftExperiments),
		IncludeDraftExperimentRefs:    optBool(m.IncludeDraftExperimentRefs),
		IncludeExperimentNames:        optBool(m.IncludeExperimentNames),
		IncludeRedirectExperiments:    optBool(m.IncludeRedirectExperiments),
		IncludeRuleIDs:                optBool(m.IncludeRuleIDs),
		IncludeProjectIDInMetadata:    optBool(m.IncludeProjectIDInMetadata),
		IncludeCustomFieldsInMetadata: optBool(m.IncludeCustomFieldsInMetadata),
		AllowedCustomFieldsInMetadata: diags.strings(ctx, m.AllowedCustomFieldsInMetadata),
		IncludeTagsInMetadata:         optBool(m.IncludeTagsInMetadata),
		ProxyEnabled:                  optBool(m.ProxyEnabled),
		ProxyHost:                     optString(m.ProxyHost),
		HashSecureAttributes:          optBool(m.HashSecureAttributes),
		RemoteEvalEnabled:             optBool(m.RemoteEvalEnabled),
		SavedGroupReferencesEnabled:   optBool(m.SavedGroupReferencesEnabled),
	}
}

func (r *sdkConnectionResource) apply(state *sdkConnectionResourceModel, c *client.SDKConnection) {
	state.ID = types.StringValue(c.ID)
	state.Name = types.StringValue(c.Name)
	state.Languages = sliceToStringList(c.Languages)
	if len(c.Languages) > 0 {
		state.Language = types.StringValue(c.Languages[0])
	}
	state.SDKVersion = types.StringValue(c.SDKVersion)
	state.Environment = types.StringValue(c.Environment)
	state.Projects = sliceToStringList(c.Projects)
	state.EncryptPayload = types.BoolValue(c.EncryptPayload)
	state.IncludeVisualExperiments = types.BoolValue(c.IncludeVisualExperiments)
	state.IncludeDraftExperiments = types.BoolValue(c.IncludeDraftExperiments)
	state.IncludeDraftExperimentRefs = types.BoolValue(c.IncludeDraftExperimentRefs)
	state.IncludeExperimentNames = types.BoolValue(c.IncludeExperimentNames)
	state.IncludeRedirectExperiments = types.BoolValue(c.IncludeRedirectExperiments)
	state.IncludeRuleIDs = types.BoolValue(c.IncludeRuleIDs)
	state.IncludeProjectIDInMetadata = types.BoolValue(c.IncludeProjectIDInMetadata)
	state.IncludeCustomFieldsInMetadata = types.BoolValue(c.IncludeCustomFieldsInMetadata)
	state.AllowedCustomFieldsInMetadata = sliceToStringList(c.AllowedCustomFieldsInMetadata)
	state.IncludeTagsInMetadata = types.BoolValue(c.IncludeTagsInMetadata)
	state.ProxyEnabled = types.BoolValue(c.ProxyEnabled)
	state.ProxyHost = types.StringValue(c.ProxyHost)
	state.HashSecureAttributes = types.BoolValue(c.HashSecureAttributes)
	state.RemoteEvalEnabled = types.BoolValue(c.RemoteEvalEnabled)
	state.SavedGroupReferencesEnabled = types.BoolValue(c.SavedGroupReferencesEnabled)
	state.Key = types.StringValue(c.Key)
	state.EncryptionKey = types.StringValue(c.EncryptionKey)
	state.ProxySigningKey = types.StringValue(c.ProxySigningKey)
	state.SSEEnabled = types.BoolValue(c.SSEEnabled)
	state.Organization = types.StringValue(c.Organization)
}

func (r *sdkConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sdkConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateSDKConnection(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create SDK connection", err.Error())
		return
	}
	r.apply(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sdkConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sdkConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	conn, err := r.client.GetSDKConnection(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read SDK connection", err.Error())
		return
	}
	r.apply(&state, conn)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *sdkConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan sdkConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state sdkConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	da := &diagAppender{diags: &resp.Diagnostics}
	in := plan.toInput(ctx, da)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateSDKConnection(ctx, state.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update SDK connection", err.Error())
		return
	}
	r.apply(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sdkConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sdkConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteSDKConnection(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete SDK connection", err.Error())
	}
}

func (r *sdkConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
