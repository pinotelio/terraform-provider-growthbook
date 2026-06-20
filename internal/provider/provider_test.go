package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// TestProviderSchema exercises GetProviderSchema, which forces the framework to
// build and validate the schema of the provider and every registered resource
// and data source. Any malformed schema (e.g. an attribute that is both
// Required and Optional, a collection attribute missing its element type, or a
// duplicate type name) surfaces here as an error diagnostic.
func TestProviderSchema(t *testing.T) {
	ctx := context.Background()
	server := providerserver.NewProtocol6(New("test")())()

	resp, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("GetProviderSchema returned an unexpected error: %s", err)
	}
	for _, d := range resp.Diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			t.Errorf("schema diagnostic: %s: %s", d.Summary, d.Detail)
		}
	}

	if len(resp.ResourceSchemas) == 0 {
		t.Error("expected at least one resource schema")
	}
	if len(resp.DataSourceSchemas) == 0 {
		t.Error("expected at least one data source schema")
	}

	// Spot-check that the headline resources are registered.
	for _, name := range []string{
		"growthbook_project",
		"growthbook_environment",
		"growthbook_feature",
		"growthbook_sdk_connection",
		"growthbook_fact_table",
		"growthbook_team",
	} {
		if _, ok := resp.ResourceSchemas[name]; !ok {
			t.Errorf("resource %q not registered", name)
		}
	}
}
