---
page_title: "emma_spots Data Source - emma"
subcategory: ""
description: |-
  Returns a list of spot instances in the project, with optional filtering by project ID.
---

# emma_spots (Data Source)

Returns a list of spot instances running in the Emma platform. Use this data source to enumerate existing spot VMs, check their status and GPU type, or reference spot IDs in other resources.

Spot instances are preemptible compute instances offered at reduced cost. This data source exposes the same placement and hardware details as `emma_vms`, plus the GPU accelerator type when applicable.

All filter attributes are optional — omit `project_id` to return spot instances across all projects your credentials can access.

## Example Usage

### List all spot instances

```terraform
data "emma_spots" "all" {}

output "spot_names" {
  value = [for s in data.emma_spots.all.spots : s.name]
}
```

### List GPU spot instances in a specific project

```terraform
data "emma_spots" "project" {
  project_id = 1297
}

output "gpu_spots" {
  value = [
    for s in data.emma_spots.project.spots : s
    if s.accelerator_type != null && s.accelerator_type != ""
  ]
}
```

### Find all active spots on a given provider

```terraform
data "emma_spots" "all" {}

output "aws_active_spots" {
  value = [
    for s in data.emma_spots.all.spots : s
    if s.provider_name == "AWS" && s.status == "active"
  ]
}
```

## Schema

### Optional

- `project_id` (Number) — Filter spot instances by project ID. Omit to return spot instances across all accessible projects.

### Read-Only

- `spots` (List of Object) — List of matching spot instances. Each element has:
  - `id` (Number) — ID of the spot instance
  - `name` (String) — Name of the spot instance
  - `status` (String) — Current status of the spot instance (e.g. "active", "poweredOff")
  - `project_id` (Number) — Project ID the spot instance belongs to
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
  - `cloud_network_type` (String) — Cloud network type of the spot instance (e.g. "default", "isolated", "multi-cloud")
  - `accelerator_type` (String) — GPU accelerator type name (e.g. "NVIDIA T4"). Empty string when no GPU is attached.
  - `created_at` (String) — Date and time when the spot instance was created (ISO 8601)
