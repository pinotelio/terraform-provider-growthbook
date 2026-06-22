package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

// clientFromProviderData extracts the configured *client.Client from the
// framework's provider data. It returns nil (without adding a diagnostic) when
// providerData is nil, which happens during early framework lifecycle calls.
func clientFromProviderData(providerData any, diags *diag.Diagnostics) *client.Client {
	if providerData == nil {
		return nil
	}
	c, ok := providerData.(*client.Client)
	if !ok {
		diags.AddError(
			"Unexpected provider data",
			"Expected *client.Client. This is a bug in the provider; please report it.",
		)
		return nil
	}
	return c
}

// diagAppender bundles a *diag.Diagnostics so resource models can convert
// nested/list attributes without threading the diagnostics object through every
// helper call.
type diagAppender struct {
	diags *diag.Diagnostics
}

// strings converts a Terraform string list/set into a Go slice, appending any
// conversion diagnostics.
func (d *diagAppender) strings(ctx context.Context, list types.List) []string {
	return stringListToSlice(ctx, list, d.diags)
}

// ---- value conversion helpers -------------------------------------------------

// optString returns nil when the attribute is null/unknown, otherwise a pointer
// to its value. Useful for building API request bodies where an omitted field
// must be left out of the JSON.
func optString(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

func optBool(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	b := v.ValueBool()
	return &b
}

func optInt64(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := v.ValueInt64()
	return &i
}

func optFloat64(v types.Float64) *float64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	f := v.ValueFloat64()
	return &f
}

// stringPtrValue maps a *string to a Terraform string, using null for nil.
func stringPtrValue(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

func boolPtrValue(b *bool) types.Bool {
	if b == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*b)
}

func int64PtrValue(i *int64) types.Int64 {
	if i == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*i)
}

func float64PtrValue(f *float64) types.Float64 {
	if f == nil {
		return types.Float64Null()
	}
	return types.Float64Value(*f)
}

// ---- string list helpers ------------------------------------------------------

// stringListToSlice converts a Terraform list/set of strings into a Go slice.
// A null or unknown list yields a nil slice.
func stringListToSlice(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var out []string
	diags.Append(list.ElementsAs(ctx, &out, false)...)
	return out
}

// sliceToStringList converts a Go slice into a Terraform list of strings. Both a
// nil and an empty slice yield a null list. GrowthBook frequently echoes unset
// optional arrays back as `[]`; collapsing empty to null keeps an omitted
// (null) configuration consistent with what the API returns, avoiding
// "inconsistent result after apply" errors and perpetual null⇄[] diffs on
// optional list attributes.
func sliceToStringList(s []string) types.List {
	if len(s) == 0 {
		return types.ListNull(types.StringType)
	}
	elems := make([]attr.Value, 0, len(s))
	for _, v := range s {
		elems = append(elems, types.StringValue(v))
	}
	list, _ := types.ListValue(types.StringType, elems)
	return list
}

// sliceToStringListEmpty is like sliceToStringList but maps a nil/empty slice to
// an empty list rather than null. Use it for Optional+Computed list attributes
// where the configuration may carry an explicit `[]`: collapsing that to null
// would make the applied value differ from the planned value ("inconsistent
// result after apply") and churn between `[]` and null on every plan.
func sliceToStringListEmpty(s []string) types.List {
	elems := make([]attr.Value, 0, len(s))
	for _, v := range s {
		elems = append(elems, types.StringValue(v))
	}
	list, _ := types.ListValue(types.StringType, elems)
	return list
}
