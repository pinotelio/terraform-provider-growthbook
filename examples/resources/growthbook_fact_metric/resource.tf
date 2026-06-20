# A proportion metric: fraction of users with at least one order.
resource "growthbook_fact_metric" "purchase_rate" {
  name        = "Purchase rate"
  metric_type = "proportion"
  datasource  = "ds_warehouse"
  projects    = [growthbook_project.web.id]

  numerator = {
    fact_table_id = growthbook_fact_table.orders.id
    column        = "$$distinctUsers"
  }
}

# A ratio metric: average order amount per purchasing user.
resource "growthbook_fact_metric" "avg_order_value" {
  name        = "Average order value"
  metric_type = "ratio"
  datasource  = "ds_warehouse"

  numerator = {
    fact_table_id = growthbook_fact_table.orders.id
    column        = "amount"
    aggregation   = "sum"
  }

  denominator = {
    fact_table_id = growthbook_fact_table.orders.id
    column        = "$$distinctUsers"
  }
}
