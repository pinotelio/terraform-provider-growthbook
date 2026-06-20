<div align="center">
  <img src="https://github.com/growthbook.png" alt="GrowthBook Logo" width="120"/>
</div>

# Terraform Provider for GrowthBook

[![Tests](https://github.com/pinotelio/terraform-provider-growthbook/actions/workflows/test.yml/badge.svg)](https://github.com/pinotelio/terraform-provider-growthbook/actions/workflows/test.yml)
[![Lint](https://github.com/pinotelio/terraform-provider-growthbook/actions/workflows/lint.yml/badge.svg)](https://github.com/pinotelio/terraform-provider-growthbook/actions/workflows/lint.yml)
[![Release](https://github.com/pinotelio/terraform-provider-growthbook/actions/workflows/release.yml/badge.svg)](https://github.com/pinotelio/terraform-provider-growthbook/actions/workflows/release.yml)
[![License: MPL-2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)
[![Terraform Registry](https://img.shields.io/badge/terraform-registry-623CE4?logo=terraform)](https://registry.terraform.io/providers/pinotelio/growthbook)
[![Go Version](https://img.shields.io/github/go-mod/go-version/pinotelio/terraform-provider-growthbook)](https://go.dev)

Manage [GrowthBook](https://www.growthbook.io) feature flagging, experimentation,
and analytics configuration as code. This provider targets the GrowthBook REST
API (`/api/v1`) and works with both GrowthBook Cloud and self-hosted instances.

## Coverage

Broad, declarative coverage of GrowthBook's configurable surface — **20 resources**
and **23 data sources**.

| Domain | Resources | |
|---|---|:--:|
| Projects & environments | `project`, `environment` | ✅ |
| Feature flags | `feature` *(per-environment enablement + force/rollout/experiment/experiment-ref rules)* | ✅ |
| SDK delivery | `sdk_connection` | ✅ |
| Targeting | `attribute`, `saved_group`, `namespace` | ✅ |
| Metrics & data modeling | `metric`, `metric_group`, `fact_table`, `fact_table_filter`, `fact_metric`, `dimension`, `segment` | ✅ |
| Experimentation & governance | `experiment_template`, `ramp_schedule_template`, `archetype`, `custom_field`, `team`, `dashboard` | ✅ |
| Data sources | one per resource above, plus `projects`, `environments`, and read-only `members` | ✅ |
| Import support | all resources support `terraform import` | ✅ |

Intentionally **out of scope** — operations that don't fit Terraform's declarative
lifecycle:

| Not managed (by design) | |
|---|:--:|
| Imperative actions — feature/experiment toggle, start, stop, snapshot | ❌ |
| Revision & review workflows (features, saved groups) | ❌ |
| Ramp-schedule runtime ops & visual editor (incl. AI) | ❌ |
| Analytics output — experiment results, queries, usage, settings | ❌ |
| Experiments, organizations, warehouse data sources, bulk import | ❌ |

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0 (or
  [OpenTofu](https://opentofu.org) >= 1.6)
- A GrowthBook account and an API key with appropriate permissions
- [Go](https://go.dev) >= 1.25 (only to build the provider from source)

## Using the provider

```hcl
terraform {
  required_providers {
    growthbook = {
      source  = "pinotelio/growthbook"
      version = "~> 0.1"
    }
  }
}

provider "growthbook" {
  # api_key  = "secret_..."                      # or GROWTHBOOK_API_KEY
  # api_url  = "https://api.growthbook.io/api/v1" # or GROWTHBOOK_API_URL
}

resource "growthbook_project" "web" {
  name        = "Web"
  description = "Web application feature flags"
}

resource "growthbook_environment" "production" {
  id            = "production"
  description   = "Production environment"
  default_state = true
  projects      = [growthbook_project.web.id]
}

resource "growthbook_feature" "new_checkout" {
  id            = "new-checkout"
  value_type    = "boolean"
  default_value = "false"
  project       = growthbook_project.web.id
  tags          = ["checkout"]

  environments = {
    production = {
      enabled = true
      rules = [
        {
          type      = "force"
          value     = "true"
          condition = jsonencode({ country = "US" })
        },
      ]
    }
  }
}
```

### Provider configuration

| Argument | Env var | Default | Description |
|---|---|---|---|
| `api_key` | `GROWTHBOOK_API_KEY` | — | API secret key (required). |
| `api_url` | `GROWTHBOOK_API_URL` | `https://api.growthbook.io/api/v1` | Base API URL including `/api/v1`. |
| `http_timeout_seconds` | — | `60` | Per-request HTTP timeout. |
| `insecure_skip_verify` | — | `false` | Skip TLS verification (self-signed self-hosted only). |
| `retry_max_attempts` | — | `4` | Retries for HTTP 429/5xx. |
| `retry_min_backoff_ms` / `retry_max_backoff_ms` | — | `500` / `5000` | Backoff bounds. |
| `page_limit` | — | `100` | Page size for list/data-source reads. |

Full resource and data-source documentation lives under [`docs/`](./docs) and on
the Terraform Registry.

## Developing the provider

```bash
make build      # build the provider binary
make test       # unit tests
make vet        # go vet
make lint       # golangci-lint
make docs       # regenerate docs/ via tfplugindocs
```

### Running locally against a build

Point Terraform at your local build with a `dev_overrides` block in `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "pinotelio/growthbook" = "/path/to/your/GOBIN"
  }
  direct {}
}
```

Then `go install .` and run `terraform plan` in an example directory (no `init`
needed under dev overrides).

### Tests

Unit tests (including a schema-validation test that exercises every resource and
data source) run with `make test` and require no network access.

The `make testacc` target is reserved for acceptance tests that run against a
live GrowthBook instance and create real resources, gated behind `TF_ACC`:

```bash
export GROWTHBOOK_API_URL="http://localhost:3100/api/v1"
export GROWTHBOOK_API_KEY="<key>"
make testacc
```

## Publishing a release

Releases are produced by [GoReleaser](https://goreleaser.com) and signed with GPG
for the Terraform Registry. Tag a semver release and the `release` workflow builds
and signs the artifacts:

```bash
git tag v0.1.0 && git push origin v0.1.0
```

Publishing to the Registry additionally requires a one-time setup: registering the
namespace on the Terraform Registry and adding your GPG public key. The repository
secrets `GPG_PRIVATE_KEY` and `GPG_PASSPHRASE` must be configured for the release
workflow.

## License

[MPL-2.0](./LICENSE).
