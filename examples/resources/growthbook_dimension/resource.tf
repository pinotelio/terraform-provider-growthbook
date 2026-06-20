resource "growthbook_dimension" "signup_source" {
  name            = "Signup source"
  datasource_id   = "ds_warehouse"
  identifier_type = "user_id"
  query           = "SELECT user_id, signup_source AS value FROM users"
  owner           = "data-team"
}
