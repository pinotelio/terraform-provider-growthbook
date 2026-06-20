# A "list" saved group: an explicit set of attribute values.
resource "growthbook_saved_group" "beta_users" {
  name          = "Beta users"
  type          = "list"
  attribute_key = "id"
  values        = ["user_1", "user_2", "user_3"]
  owner         = "growth-team"
}

# A "condition" saved group: a reusable targeting condition.
resource "growthbook_saved_group" "us_enterprise" {
  name      = "US enterprise accounts"
  type      = "condition"
  condition = jsonencode({ country = "US", plan = "enterprise" })
}
