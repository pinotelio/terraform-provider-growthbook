package client

import (
	"encoding/json"
	"strings"
	"testing"
)

// Issue 02: an environment with no targeting rules must serialize `rules` as an
// empty array, never omit the key (GrowthBook validates it as a required array).
func TestFeatureEnvironmentInput_EmptyRulesSerializeAsArray(t *testing.T) {
	enabled := true
	in := FeatureCreateInput{
		ID:        "repro-empty-rules",
		ValueType: "boolean",
		Environments: map[string]FeatureEnvironmentInput{
			"production": {
				Enabled: &enabled,
				Rules:   make([]FeatureRule, 0), // explicitly empty
			},
		},
	}

	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %s", err)
	}
	got := string(b)
	if !strings.Contains(got, `"rules":[]`) {
		t.Errorf("expected an empty rules array in payload, got: %s", got)
	}
}

// Issue 03: the update payload must not include the create-only `valueType`
// field, which the GrowthBook update endpoint rejects as an unrecognized key.
func TestFeatureUpdateInput_OmitsValueType(t *testing.T) {
	desc := "updated"
	in := FeatureUpdateInput{
		Description: &desc,
		Environments: map[string]FeatureEnvironmentInput{
			"production": {Rules: make([]FeatureRule, 0)},
		},
	}

	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %s", err)
	}
	if got := string(b); strings.Contains(got, "valueType") {
		t.Errorf("update payload must not contain valueType, got: %s", got)
	}
}
