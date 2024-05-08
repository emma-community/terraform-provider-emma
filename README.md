# Custom Terraform Provider for Emma

## Overview

This Terraform provider allows you to manage resources within your Emma infrastructure. By using Terraform, you can 
define your infrastructure as code and easily provision, update, and manage resources in a repeatable and automated manner.

## Features

- Provision and manage virtual machines.
- Utilize spot instances for cost-effective computing.
- Manage SSH keys for secure access to instances.
- Define and manage security groups to control network traffic.

## Installation

1. **Prerequisites**: Ensure you have Terraform installed on your system. You can download it from the [Terraform website](https://developer.hashicorp.com/terraform/install).
2. **Download the Provider**: Download the latest release of the provider binary from the releases page.
3. **Install the Provider**: Move the downloaded binary to the Terraform plugins directory:
   ```bash
   mv terraform-provider-emma ~/.terraform.d/plugins/
    ```

## Usage

1. **Define Provider Configuration**: Add the provider configuration to your Terraform configuration file (e.g., main.tf):
   ```hcl
   provider "emma" {
        client_id     = "your client id"
        client_secret = "your client secret"
   }
   ```

2. **Define Resources**: Define the resources you want to manage in your Terraform configuration. Here's an example 
of provisioning a virtual machine:
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

3. **Run Terraform Commands**: Use Terraform commands (`terraform init`, `terraform plan`, `terraform apply`, etc.) 
to apply your configuration and manage your infrastructure.

## Authentication

To authenticate with Emma's infrastructure, provide the necessary credentials using the `client_id` and `client_secret` 
options in your provider configuration.
