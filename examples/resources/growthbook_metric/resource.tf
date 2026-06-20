resource "growthbook_metric" "signup" {
  name          = "Signups"
  type          = "binomial"
  datasource_id = "ds_warehouse"
  description   = "User completed signup"
  projects      = [growthbook_project.web.id]
  tags          = ["acquisition"]
}
