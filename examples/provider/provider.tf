terraform {
  required_providers {
    growthbook = {
      source  = "pinotelio/growthbook"
      version = "~> 0.1"
    }
  }
}

# Credentials can also be supplied via GROWTHBOOK_API_KEY / GROWTHBOOK_API_URL.
provider "growthbook" {
  api_key = var.growthbook_api_key
  api_url = "https://api.growthbook.io/api/v1"
}
