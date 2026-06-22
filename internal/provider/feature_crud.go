package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

// applyComputed copies the server-managed scalar fields of a feature into the
// model. Environment configuration is intentionally not touched here so that
// Create/Update can preserve the practitioner's configured value verbatim.
func (r *featureResource) applyComputed(m *featureResourceModel, f *client.Feature) {
	m.ID = types.StringValue(f.ID)
	m.Description = types.StringValue(f.Description)
	m.Owner = types.StringValue(f.Owner)
	m.Project = types.StringValue(f.Project)
	m.Archived = types.BoolValue(f.Archived)
	m.ValueType = types.StringValue(f.ValueType)
	m.DefaultValue = types.StringValue(f.DefaultValue)
	// tags is Optional+Computed: keep an explicit empty list as `[]` (not null)
	// so a configured `tags = []` does not produce an inconsistent result or a
	// perpetual `[] -> null` diff.
	m.Tags = sliceToStringListEmpty(f.Tags)
	m.Prerequisites = sliceToStringList(f.Prerequisites)
	m.DateCreated = types.StringValue(f.DateCreated)
	m.DateUpdated = types.StringValue(f.DateUpdated)
}

func (r *featureResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan featureResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.FeatureCreateInput{
		ID:           plan.ID.ValueString(),
		Archived:     optBool(plan.Archived),
		Description:  optString(plan.Description),
		Owner:        optString(plan.Owner),
		Project:      optString(plan.Project),
		ValueType:    plan.ValueType.ValueString(),
		DefaultValue: optString(plan.DefaultValue),
		Tags:         stringListToSlice(ctx, plan.Tags, &resp.Diagnostics),
		Environments: buildEnvironments(ctx, plan.Environments, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateFeature(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create feature", err.Error())
		return
	}
	r.applyComputed(&plan, created)
	// Environments keep their planned value to avoid post-apply inconsistency.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *featureResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state featureResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	feature, err := r.client.GetFeature(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read feature", err.Error())
		return
	}

	r.applyComputed(&state, feature)
	// Environments are write-only: the practitioner's configured value is
	// preserved verbatim rather than re-read from the API. GrowthBook populates
	// server-side defaults on rules (e.g. `coverage = 1`, `enabled = true`, a
	// generated rule `id`) that the configuration never sets; reflecting those
	// back into state would produce a perpetual diff that can never converge.
	// On import there is no prior configuration, so seed environments from the
	// API to give the practitioner a starting point.
	if len(state.Environments) == 0 {
		state.Environments = flattenEnvironments(feature)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *featureResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan featureResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.FeatureUpdateInput{
		Archived:     optBool(plan.Archived),
		Description:  optString(plan.Description),
		Owner:        optString(plan.Owner),
		Project:      optString(plan.Project),
		DefaultValue: optString(plan.DefaultValue),
		Tags:         stringListToSlice(ctx, plan.Tags, &resp.Diagnostics),
		Environments: buildEnvironments(ctx, plan.Environments, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateFeature(ctx, plan.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update feature", err.Error())
		return
	}
	r.applyComputed(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *featureResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state featureResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteFeature(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete feature", err.Error())
	}
}
