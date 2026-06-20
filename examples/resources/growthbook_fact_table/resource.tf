resource "growthbook_fact_table" "orders" {
  name          = "Orders"
  description   = "One row per completed order"
  datasource    = "ds_warehouse"
  user_id_types = ["user_id"]
  sql           = "SELECT user_id, timestamp, amount, country FROM orders"
  projects      = [growthbook_project.web.id]

  columns = [
    {
      column   = "amount"
      datatype = "number"
      name     = "Order amount"
    },
    {
      column   = "country"
      datatype = "string"
    },
  ]
}
