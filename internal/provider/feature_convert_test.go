package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Issue 02: an environment whose rules list is empty must produce a non-nil
// slice so it serializes as `[]` rather than being dropped from the request.
func TestBuildEnvironments_EmptyRulesIsNonNil(t *testing.T) {
	var diags diag.Diagnostics
	envs := map[string]featureEnvModel{
		"production": {Enabled: types.BoolValue(true), Rules: nil},
	}

	out := buildEnvironments(context.Background(), envs, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	env, ok := out["production"]
	if !ok {
		t.Fatal("production environment missing from output")
	}
	if env.Rules == nil {
		t.Fatal("expected a non-nil (empty) rules slice so it marshals as []")
	}
	if len(env.Rules) != 0 {
		t.Fatalf("expected zero rules, got %d", len(env.Rules))
	}
}

// Issue 04: tags must collapse an empty slice to an empty list (not null) so a
// configured `tags = []` doesn't yield an inconsistent result or a perpetual
// `[] -> null` diff.
func TestSliceToStringListEmpty_EmptyIsNotNull(t *testing.T) {
	list := sliceToStringListEmpty(nil)
	if list.IsNull() {
		t.Fatal("expected an empty (non-null) list for a nil slice")
	}
	if n := len(list.Elements()); n != 0 {
		t.Fatalf("expected zero elements, got %d", n)
	}

	list = sliceToStringListEmpty([]string{"a", "b"})
	if list.IsNull() || len(list.Elements()) != 2 {
		t.Fatalf("expected a 2-element list, got null=%v len=%d", list.IsNull(), len(list.Elements()))
	}
}
