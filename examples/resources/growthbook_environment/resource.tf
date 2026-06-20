resource "growthbook_environment" "production" {
  id             = "production"
  description    = "Production environment"
  default_state  = true
  toggle_on_list = true
  projects       = [growthbook_project.web.id]
}
