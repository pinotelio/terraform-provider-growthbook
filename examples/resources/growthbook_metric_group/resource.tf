resource "growthbook_metric_group" "activation" {
  name        = "Activation metrics"
  description = "Core activation funnel metrics"
  datasource  = "ds_warehouse"
  metrics     = [growthbook_metric.signup.id]
  tags        = ["activation"]
}
