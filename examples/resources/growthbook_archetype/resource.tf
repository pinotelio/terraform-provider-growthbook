resource "growthbook_archetype" "power_user" {
  name        = "Power user"
  description = "Representative attribute set for a power user"
  is_public   = true
  attributes  = jsonencode({ plan = "enterprise", country = "US", logins = 100 })
  projects    = [growthbook_project.web.id]
}
