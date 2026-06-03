---
page_title: "emma_volumes Data Source - emma"
subcategory: ""
description: |-
  Returns a list of all storage volumes in the project.
---

# emma_volumes (Data Source)

Returns a list of all storage volumes in the Emma platform. Use this data source to enumerate both system and data volumes, inspect their attachment status, or find volume IDs to reference in other resources.

This data source returns all volumes accessible to your credentials and does not accept filter arguments — use Terraform's `for` expressions to filter the result list in your configuration.

## Example Usage

### List all volumes

```terraform
data "emma_volumes" "all" {}

output "volume_names" {
  value = [for v in data.emma_volumes.all.volumes : v.name]
}
```

### Find all unattached data volumes

```terraform
data "emma_volumes" "all" {}

output "free_volumes" {
  value = [
    for v in data.emma_volumes.all.volumes : v
    if !v.is_system && v.attached_to_id == null
  ]
}
```

### Inspect volumes in a specific data center

```terraform
data "emma_volumes" "all" {}

output "eu_west_volumes" {
  value = [
    for v in data.emma_volumes.all.volumes : v
    if v.data_center_id == "aws-eu-west-1"
  ]
}
```

## Schema

### Read-Only

- `volumes` (List of Object) — List of all accessible storage volumes. Each element has:
  - `id` (String) — Volume ID
  - `name` (String) — Volume name
  - `size_gb` (Number) — Volume size in gigabytes
  - `volume_type` (String) — Volume type (e.g. "ssd", "hdd")
  - `is_system` (Boolean) — Whether the volume contains the operating system
  - `status` (String) — Current status of the volume (e.g. "active", "available")
  - `attached_to_id` (Number) — ID of the compute instance the volume is attached to. Null when the volume is unattached.
  - `project_id` (Number) — Project ID that owns the volume
  - `data_center_id` (String) — Data center ID where the volume is located
  - `created_at` (String) — Creation timestamp (ISO 8601)
  - `created_by_name` (String) — Name of the user who created the volume
  - `created_by_id` (Number) — ID of the user who created the volume
  - `modified_at` (String) — Date and time when the volume was last modified (ISO 8601)
  - `modified_by_name` (String) — Name of the user who last modified the volume
  - `modified_by_id` (Number) — ID of the user who last modified the volume
