//go:build tools

// Package tools pins build-time tooling so `go generate` uses a consistent
// version of the documentation generator. It is never compiled into the
// provider binary (guarded by the `tools` build tag).
package tools

import (
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
