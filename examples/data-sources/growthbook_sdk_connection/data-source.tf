data "growthbook_sdk_connection" "web_prod" {
  id = "sdk-abc123"
}

output "sdk_client_key" {
  value = data.growthbook_sdk_connection.web_prod.key
}
