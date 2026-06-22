// Package provider implements the GrowthBook Terraform provider using the
// terraform-plugin-framework.
package provider

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-growthbook/internal/client"
)

// Ensure the provider satisfies the framework interface.
var _ provider.Provider = (*growthbookProvider)(nil)

// growthbookProvider is the provider implementation.
type growthbookProvider struct {
	version string
}

// New returns a provider factory bound to the given build version.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &growthbookProvider{version: version}
	}
}

type providerModel struct {
	APIKey             types.String `tfsdk:"api_key"`
	APIURL             types.String `tfsdk:"api_url"`
	HTTPTimeoutSeconds types.Int64  `tfsdk:"http_timeout_seconds"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
	RetryMaxAttempts   types.Int64  `tfsdk:"retry_max_attempts"`
	RetryMinBackoffMs  types.Int64  `tfsdk:"retry_min_backoff_ms"`
	RetryMaxBackoffMs  types.Int64  `tfsdk:"retry_max_backoff_ms"`
	MaxRequestsPerMin  types.Int64  `tfsdk:"max_requests_per_minute"`
	PageLimit          types.Int64  `tfsdk:"page_limit"`
}

func (p *growthbookProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "growthbook"
	resp.Version = p.version
}

func (p *growthbookProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage [GrowthBook](https://www.growthbook.io) feature flags, " +
			"experimentation, and analytics configuration as code.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Description: "API key (secret) used to authenticate against the GrowthBook REST API. " +
					"May also be provided via the `GROWTHBOOK_API_KEY` environment variable.",
			},
			"api_url": schema.StringAttribute{
				Optional: true,
				Description: "Base URL of the GrowthBook REST API, including the `/api/v1` suffix. " +
					"Defaults to `https://api.growthbook.io/api/v1`. May also be provided via the " +
					"`GROWTHBOOK_API_URL` environment variable.",
			},
			"http_timeout_seconds": schema.Int64Attribute{
				Optional:    true,
				Description: "Per-request HTTP timeout in seconds. Defaults to 60.",
			},
			"insecure_skip_verify": schema.BoolAttribute{
				Optional: true,
				Description: "Disable TLS certificate verification when calling the API. " +
					"Only useful for self-hosted instances with self-signed certificates; not recommended for production.",
			},
			"retry_max_attempts": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum number of attempts for retryable (HTTP 429/5xx) responses. Defaults to 4.",
			},
			"retry_min_backoff_ms": schema.Int64Attribute{
				Optional:    true,
				Description: "Minimum backoff between retries, in milliseconds. Defaults to 500.",
			},
			"retry_max_backoff_ms": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum backoff between retries, in milliseconds. Defaults to 5000.",
			},
			"max_requests_per_minute": schema.Int64Attribute{
				Optional: true,
				Description: "Client-side request rate limit. When set, requests are evenly " +
					"spaced so no more than this many are sent per minute, preventing large " +
					"plans/refreshes from tripping the GrowthBook API rate limit (HTTP 429). " +
					"Set it at or below your instance's limit (self-hosted defaults to 60). " +
					"Defaults to 0 (disabled).",
			},
			"page_limit": schema.Int64Attribute{
				Optional:    true,
				Description: "Page size used when reading paginated list endpoints. Defaults to 100.",
			},
		},
	}
}

func (p *growthbookProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := stringWithEnv(cfg.APIKey, "GROWTHBOOK_API_KEY")
	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing GrowthBook API key",
			"Set the `api_key` provider argument or the GROWTHBOOK_API_KEY environment variable.",
		)
		return
	}

	apiURL := stringWithEnv(cfg.APIURL, "GROWTHBOOK_API_URL")

	timeout := int64OrDefault(cfg.HTTPTimeoutSeconds, 60)
	insecure := cfg.InsecureSkipVerify.ValueBool()
	httpClient := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}, //nolint:gosec // opt-in via insecure_skip_verify
		},
	}

	retry := client.RetryPolicy{
		MaxAttempts: int(int64OrDefault(cfg.RetryMaxAttempts, 4)),
		MinBackoff:  time.Duration(int64OrDefault(cfg.RetryMinBackoffMs, 500)) * time.Millisecond,
		MaxBackoff:  time.Duration(int64OrDefault(cfg.RetryMaxBackoffMs, 5000)) * time.Millisecond,
		Multiplier:  2.0,
	}

	c := client.New(apiURL, apiKey,
		client.WithHTTPClient(httpClient),
		client.WithRetryPolicy(retry),
		client.WithRateLimit(int(int64OrDefault(cfg.MaxRequestsPerMin, 0))),
		client.WithPageLimit(int(int64OrDefault(cfg.PageLimit, 100))),
	)

	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *growthbookProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newProjectResource,
		newEnvironmentResource,
		newFeatureResource,
		newSDKConnectionResource,
		newAttributeResource,
		newSavedGroupResource,
		newNamespaceResource,
		newMetricResource,
		newFactTableResource,
		newFactTableFilterResource,
		newFactMetricResource,
		newMetricGroupResource,
		newDimensionResource,
		newSegmentResource,
		newExperimentTemplateResource,
		newRampScheduleTemplateResource,
		newArchetypeResource,
		newCustomFieldResource,
		newTeamResource,
		newDashboardResource,
	}
}

func (p *growthbookProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newProjectDataSource,
		newProjectsDataSource,
		newEnvironmentDataSource,
		newEnvironmentsDataSource,
		newFeatureDataSource,
		newSDKConnectionDataSource,
		newAttributeDataSource,
		newSavedGroupDataSource,
		newNamespaceDataSource,
		newMetricDataSource,
		newFactTableDataSource,
		newFactMetricDataSource,
		newMetricGroupDataSource,
		newDimensionDataSource,
		newSegmentDataSource,
		newFactTableFilterDataSource,
		newExperimentTemplateDataSource,
		newRampScheduleTemplateDataSource,
		newArchetypeDataSource,
		newCustomFieldDataSource,
		newTeamDataSource,
		newDashboardDataSource,
		newMembersDataSource,
	}
}

// stringWithEnv resolves a config string, falling back to an environment
// variable when the config value is null/unknown/empty.
func stringWithEnv(v types.String, env string) string {
	if !v.IsNull() && !v.IsUnknown() && v.ValueString() != "" {
		return v.ValueString()
	}
	return os.Getenv(env)
}

func int64OrDefault(v types.Int64, def int64) int64 {
	if v.IsNull() || v.IsUnknown() {
		return def
	}
	return v.ValueInt64()
}
