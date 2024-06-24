# Terraform Provider Emma

## Overview

This [Terraform Provider Emma](https://registry.terraform.io/providers/emma-community/emma/latest) allows you to manage 
multi-cloud resources. The [emma platform](https://www.emma.ms/) empowers you to effortlessly deploy and manage cloud 
resources across diverse environments, spanning on-premises, private, and public clouds. Whether you're a seasoned cloud 
professional honing your multi-cloud setup or diving into cloud management for the first time, our cloud-agnostic 
approach guarantees freedom to leverage the right cloud services you need.

## Features

- Provision and manage virtual machines.
- Utilize spot instances for cost-effective computing.
- Manage SSH keys for secure access to instances.
- Define and manage security groups to control network traffic.

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
      ssh_key_id         = emma_ssh_key.ssh_key.id
   }
   ```

4. **Run Terraform Commands**: Use Terraform commands (`terraform plan`, `terraform apply`, etc.) 
to apply your configuration and manage your infrastructure.

## Authentication

To authenticate with Emma's infrastructure, provide the necessary credentials using the `client_id` and `client_secret` 
options in your provider configuration.
