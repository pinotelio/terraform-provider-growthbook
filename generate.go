//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name growthbook

// Package main hosts the `go generate` directive that produces the Terraform
// Registry documentation under docs/ from the provider schema and examples/.
package main
