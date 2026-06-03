---
page_title: "emma_security_groups Data Source - emma"
subcategory: ""
description: |-
  Returns a list of security groups in the project, with optional filtering by project ID.
---

# emma_security_groups (Data Source)

Returns a list of security groups defined in the Emma platform. Use this data source to look up security group IDs for attaching to compute instances, or to audit synchronization and recomposing status across providers.

All filter attributes are optional — omit `project_id` to return security groups across all projects your credentials can access.

## Example Usage

### List all security groups

```terraform
data "emma_security_groups" "all" {}

output "security_group_names" {
  value = [for sg in data.emma_security_groups.all.security_groups : sg.name]
}
```

### Find security groups in a specific project

```terraform
data "emma_security_groups" "project" {
  project_id = 1297
}

output "synced_groups" {
  value = [
    for sg in data.emma_security_groups.project.security_groups : sg
    if sg.synchronization_status == "synced"
  ]
}
```

### Resolve a security group ID by name for use in a VM resource

```terraform
data "emma_security_groups" "all" {}

locals {
  web_sg_id = one([
    for sg in data.emma_security_groups.all.security_groups : sg.id
    if sg.name == "web-tier"
  ])
}
```

## Schema

### Optional

- `project_id` (Number) — Filter security groups by project ID. Omit to return security groups across all accessible projects.

### Read-Only

- `security_groups` (List of Object) — List of matching security groups. Each element has:
  - `id` (Number) — ID of the security group
  - `name` (String) — Name of the security group
  - `created_at` (String) — Date and time when the security group was created (ISO 8601)
  - `created_by_name` (String) — Name of the user who created the security group
  - `modified_at` (String) — Date and time when the security group was last modified (ISO 8601)
  - `synchronization_status` (String) — Synchronization status of the security group rules across providers (e.g. "synced", "not_synced")
  - `recomposing_status` (String) — Recomposing status of the security group when new compute instances are added (e.g. "done", "in_progress")
