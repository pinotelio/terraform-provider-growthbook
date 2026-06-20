data "growthbook_project" "web" {
  id = "prj_abc123"
}

output "web_project_name" {
  value = data.growthbook_project.web.name
}
