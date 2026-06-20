data "growthbook_feature" "new_checkout" {
  id = "new-checkout"
}

output "checkout_default_value" {
  value = data.growthbook_feature.new_checkout.default_value
}
