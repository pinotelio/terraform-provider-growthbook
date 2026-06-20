resource "growthbook_custom_field" "jira_ticket" {
  id       = "jira-ticket"
  name     = "Jira ticket"
  type     = "text"
  required = false
  sections = ["feature", "experiment"]
}

resource "growthbook_custom_field" "risk_level" {
  id       = "risk-level"
  name     = "Risk level"
  type     = "enum"
  values   = "low,medium,high"
  required = true
  sections = ["experiment"]
}
