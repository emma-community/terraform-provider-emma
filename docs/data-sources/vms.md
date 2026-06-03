---
page_title: "emma_vms Data Source - emma"
subcategory: ""
description: |-
  Returns a list of virtual machines in the project, with optional filtering by project ID.
---

# emma_vms (Data Source)

Returns a list of virtual machines running in the Emma platform. Use this data source to enumerate existing VMs, inspect their status and placement, or reference VM IDs in other resources.

All filter attributes are optional — omit `project_id` to return VMs across all projects your credentials can access.

## Example Usage

### List all VMs in the default project

```terraform
data "emma_vms" "all" {}

output "vm_names" {
  value = [for v in data.emma_vms.all.vms : v.name]
}
```

### List VMs in a specific project

```terraform
data "emma_vms" "project" {
  project_id = 1297
}

output "running_vms" {
  value = [for v in data.emma_vms.project.vms : v if v.status == "active"]
}
```

### Find VMs running in a specific data center

```terraform
data "emma_vms" "all" {}

output "aws_east_vms" {
  value = [for v in data.emma_vms.all.vms : v if v.data_center_id == "aws-us-east-1"]
}
```

## Schema

### Optional

- `project_id` (Number) — Filter virtual machines by project ID. Omit to return VMs across all accessible projects.

### Read-Only

- `vms` (List of Object) — List of matching virtual machines. Each element has:
  - `id` (Number) — ID of the virtual machine
  - `name` (String) — Name of the virtual machine
  - `status` (String) — Current status of the virtual machine (e.g. "active", "poweredOff")
  - `project_id` (Number) — Project ID the virtual machine belongs to
  - `provider_id` (Number) — ID of the cloud provider
  - `provider_name` (String) — Name of the cloud provider (e.g. "AWS", "Azure")
  - `location_id` (Number) — ID of the geographic location
  - `location_name` (String) — Name of the geographic location
  - `data_center_id` (String) — ID of the data center (e.g. "aws-us-east-1")
  - `data_center_name` (String) — Name of the data center
  - `os_id` (Number) — ID of the operating system
  - `os_type` (String) — Type of the operating system (e.g. "Ubuntu")
  - `vcpu` (Number) — Number of virtual CPUs
  - `vcpu_type` (String) — vCPU type: "shared", "standard", or "hpc"
  - `ram_gb` (Number) — RAM in gigabytes
  - `cloud_network_type` (String) — Cloud network type of the virtual machine (e.g. "default", "isolated", "multi-cloud")
  - `created_at` (String) — Date and time when the virtual machine was created (ISO 8601)
