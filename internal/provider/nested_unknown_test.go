package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// nestedGuardChild/Parent mirror the shape used throughout the provider: a
// nested attribute whose model field is a native Go *struct.
type nestedGuardChild struct {
	Name types.String `tfsdk:"name"`
}
type nestedGuardParent struct {
	Child *nestedGuardChild `tfsdk:"child"`
}

func nestedGuardSchema(computed bool) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"child": schema.SingleNestedAttribute{
				Optional: true,
				Computed: computed,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{Optional: true, Computed: computed},
				},
			},
		},
	}
}

// TestNestedNativePointerNullDecodes verifies the pattern the provider relies
// on: an Optional (NOT Computed) nested attribute backed by a native *struct
// decodes cleanly when null. This is why native nested settings blocks must not
// be Computed — a Computed nested attribute is planned as *unknown* on
// create-when-unset, and unknown cannot be reflected into a native Go type
// (see TestNestedNativePointerUnknownFails for the failure that guards against
// reintroducing Computed on such attributes).
func TestNestedNativePointerNullDecodes(t *testing.T) {
	ctx := context.Background()
	childType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String}}
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"child": childType}}
	raw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"child": tftypes.NewValue(childType, nil), // null
	})

	plan := tfsdk.Plan{Schema: nestedGuardSchema(false), Raw: raw}
	var m nestedGuardParent
	if diags := plan.Get(ctx, &m); diags.HasError() {
		t.Fatalf("null nested object into native *struct should decode cleanly, got: %v", diags)
	}
	if m.Child != nil {
		t.Fatalf("expected nil child for null nested object, got %+v", m.Child)
	}
}

// TestNestedNativePointerUnknownFails documents (and locks in) the hazard: an
// unknown nested object cannot be decoded into a native *struct. Native nested
// attributes must therefore be Optional-only, never Computed.
func TestNestedNativePointerUnknownFails(t *testing.T) {
	ctx := context.Background()
	childType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String}}
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"child": childType}}
	raw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"child": tftypes.NewValue(childType, tftypes.UnknownValue),
	})

	plan := tfsdk.Plan{Schema: nestedGuardSchema(true), Raw: raw}
	var m nestedGuardParent
	if diags := plan.Get(ctx, &m); !diags.HasError() {
		t.Fatal("expected an error decoding an unknown nested object into a native *struct, got none")
	}
}
