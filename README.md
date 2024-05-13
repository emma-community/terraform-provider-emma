# Terraform Provider Emma

## Overview

This [Terraform Provider Emma](https://registry.terraform.io/providers/emma-community/emma/latest) allows you to manage 
resources within your Emma infrastructure. By using Terraform, you can define your infrastructure as code and easily 
provision, update, and manage resources in a repeatable and automated manner.

## Features

- Provision and manage virtual machines.

## Installation

1. **Prerequisites**: Ensure you have Terraform installed on your system. You can download it from the [Terraform website](https://developer.hashicorp.com/terraform/install).
2. **Define Provider Configuration**: To install this provider, copy and paste this code into your Terraform configuration. 
Then, run `terraform init`:
   ```hcl
   terraform {
     required_providers {
       emma = {
         source = "emma-community/emma"
         version = "0.0.1-alpha"
         }
      }
   }

   provider "emma" {
     client_id     = "your client id"
     client_secret = "your client secret"
   }
   ```

3. **Define Resources**: Define the resources you want to manage in your Terraform configuration. Here's an example 
of provisioning a virtual machine, you can find more documentation on the [terraform provider page](https://registry.terraform.io/providers/emma-community/emma/latest/docs):
   ```hcl
   resource "emma_vm" "vm" {
      name               = "Example"
      data_center_id     = data.emma_data_center.aws.id
      os_id              = data.emma_operating_system.ubuntu.id
      cloud_network_type = "multi-cloud"
      vcpu_type          = "shared"
      vcpu               = 2
      ram_gb             = 1
      volume_type        = "ssd"
      volume_gb          = 8
      ssh_key_id         = 1
   }
   ```

4. **Run Terraform Commands**: Use Terraform commands (`terraform plan`, `terraform apply`, etc.) 
to apply your configuration and manage your infrastructure.

## Authentication

To authenticate with Emma's infrastructure, provide the necessary credentials using the `client_id` and `client_secret` 
options in your provider configuration.
