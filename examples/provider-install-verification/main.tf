terraform {
  required_providers {
    emma = {
      source = "hashicorp.com/edu/emma"
    }
  }
}

# terraform {
#   required_providers {
#     emma = {
#       source = "emma-community/emma"
#       version = "0.0.1-alpha"
#     }
#   }
# }

provider "emma" {
  host          = "https://customer-gateway.dev.emma.ms"
  client_id     = "fcd4dba5-8177-4940-80b6-565726e7464f"
  client_secret = "ba2eca0d-06bc-4b79-9f90-85cd39817416"
}

data "emma_data_center" "aws" {
  name          = "eu-north-1"
  provider_name = "Amazon EC2"
}

data "emma_data_center" "aws_spot" {
  name          = "ap-southeast-4"
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

resource "emma_ssh_key" "ssh_key" {
  name     = "demo-1305"
  key_type = "RSA"
}

resource "emma_security_group" "security_group" {
  name = "demo-1305"
  rules = [
    {
      direction = "INBOUND"
      protocol  = "all"
      ports     = "8080"
      ip_range  = "68.183.5.11/32"
    },
    {
      direction = "INBOUND"
      protocol  = "all"
      ports     = "8080"
      ip_range  = "68.183.5.111/32"
    }
  ]
}


resource "emma_vm" "vm" {
  name               = "demo-1305"
  data_center_id     = data.emma_data_center.aws.id
  os_id              = data.emma_operating_system.ubuntu.id
  cloud_network_type = "multi-cloud"
  vcpu_type          = "shared"
  vcpu               = 2
  ram_gb             = 1
  volume_type        = "ssd"
  volume_gb          = 8
  security_group_id  = emma_security_group.security_group.id
  ssh_key_id =           emma_ssh_key.ssh_key.id
}

# resource "emma_vm" "vm_import" {
#   name               = "vm-default-5th3ix6f"
#   data_center_id     = "aws-ap-south-1"
#   os_id              = 34
#   cloud_network_type = "default"
#   vcpu_type          = "shared"
#   vcpu               = 2
#   ram_gb             = 1
#   volume_type        = "ssd"
#   volume_gb          = 8
#   ssh_key_id         = 651
# }

resource "emma_spot_instance" "spot_instance" {
  name               = "demo-1305"
  data_center_id     = data.emma_data_center.aws_spot.id
  os_id              = data.emma_operating_system.ubuntu.id
  cloud_network_type = "multi-cloud"
  vcpu_type          = "shared"
  vcpu               = 2
  ram_gb             = 1
  volume_type        = "ssd"
  volume_gb          = 8
  security_group_id  = emma_security_group.security_group.id
  ssh_key_id         = emma_ssh_key.ssh_key.id
  price              = 0.00305205479452
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

output "emma_ssh_key" {
  value = emma_ssh_key.ssh_key
}

output "emma_security_group" {
  value = emma_security_group.security_group
}

output "emma_vm" {
  value = emma_vm.vm
}

output "emma_spot_instance" {
  value = emma_spot_instance.spot_instance
}
