terraform {
  required_providers {
    emma = {
      source = "hashicorp.com/edu/emma"
    }
  }
}

provider "emma" {
  client_id     = "client_id"
  client_secret = "client_secret"
}

data "emma_data_center" "aws" {
  name          = "eu-north-1"
  provider_name = "Amazon EC2"
}

data "emma_location" "stockholm" {
  name = "Stockholm"
}

data "emma_operating_system" "ubuntu" {
  type         = "Ubuntu"
  architecture = "x86-64"
  version      = "22.04"
}

data "emma_provider" "aws" {
  name = "Amazon EC2"
}

resource "emma_vm" "vm" {
  name               = "demo-2604"
  data_center_id     = data.emma_data_center.aws.id
  os_id              = data.emma_operating_system.ubuntu.id
  cloud_network_type = "multi-cloud"
  vcpu_type          = "shared"
  vcpu               = 2
  ram_gb             = 1
  volume_type        = "ssd"
  volume_gb          = 8
  ssh_key_id         = 570
}

output "emma_data_center" {
  value = data.emma_data_center.aws
}

output "emma_data_location" {
  value = data.emma_location.stockholm
}

output "emma_operating_system" {
  value = data.emma_operating_system.ubuntu
}

output "emma_provider" {
  value = data.emma_provider.aws
}