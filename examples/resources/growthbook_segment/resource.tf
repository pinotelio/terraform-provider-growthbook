resource "growthbook_segment" "active_paid" {
  name            = "Active paid users"
  type            = "SQL"
  datasource_id   = "ds_warehouse"
  identifier_type = "user_id"
  sql             = "SELECT user_id, date FROM subscriptions WHERE status = 'active'"
  owner           = "data-team"
}
