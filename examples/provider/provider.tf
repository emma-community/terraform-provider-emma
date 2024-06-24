terraform {
  required_providers {
    emma = {
      source  = "emma-community/emma"
      version = "0.1.0"
    }
  }
}

provider "emma" {
  client_id     = "your client id"
  client_secret = "your client secret"
}
