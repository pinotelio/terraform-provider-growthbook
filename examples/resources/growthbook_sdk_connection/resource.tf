resource "growthbook_sdk_connection" "web_prod" {
  name        = "web-production"
  language    = "javascript"
  environment = growthbook_environment.production.id
  projects    = [growthbook_project.web.id]

  encrypt_payload            = true
  include_visual_experiments = true
}
