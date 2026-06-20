resource "growthbook_attribute" "country" {
  property       = "country"
  datatype       = "string"
  description    = "ISO country code of the user"
  hash_attribute = false
  projects       = [growthbook_project.web.id]
}

resource "growthbook_attribute" "plan" {
  property = "plan"
  datatype = "enum"
  enum     = "free,pro,enterprise"
}
