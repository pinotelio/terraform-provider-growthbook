resource "growthbook_feature" "new_checkout" {
  id            = "new-checkout"
  description   = "Roll out the redesigned checkout flow"
  value_type    = "boolean"
  default_value = "false"
  project       = growthbook_project.web.id
  tags          = ["checkout", "growth"]

  environments = {
    production = {
      enabled = true
      rules = [
        {
          type        = "force"
          description = "Force on for internal users"
          value       = "true"
          condition   = jsonencode({ email = { "$regex" = "@example\\.com$" } })
        },
        {
          type           = "rollout"
          value          = "true"
          coverage       = 0.25
          hash_attribute = "id"
        },
      ]
    }
  }
}
