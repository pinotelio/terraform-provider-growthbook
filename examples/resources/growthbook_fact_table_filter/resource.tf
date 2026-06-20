resource "growthbook_fact_table_filter" "us_only" {
  fact_table_id = growthbook_fact_table.orders.id
  name          = "US only"
  value         = "country = 'US'"
}
