resource "growthbook_project" "web" {
  name         = "Web"
  description  = "Web application feature flags and experiments"
  stats_engine = "bayesian"
}
