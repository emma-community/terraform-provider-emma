---
page_title: "emma_workflow_templates Data Source - emma"
subcategory: ""
description: |-
  Returns a list of workflow templates in the Emma platform, with optional filtering by name or status.
---

# emma_workflow_templates (Data Source)

Returns a list of workflow templates in the Emma platform. Use this data source to enumerate existing templates, look up IDs, or reference templates in other resources without importing them into Terraform state.

All filter attributes are optional — omit them to return all accessible templates.

To create and manage workflow templates as Terraform-owned infrastructure, use the [`emma_workflow_template` resource](../resources/workflow_template.md).

## Example Usage

### List all published templates

```terraform
data "emma_workflow_templates" "published" {
  status = "PUBLISHED"
}

output "template_names" {
  value = [for t in data.emma_workflow_templates.published.workflow_templates : t.name]
}
```

### Find templates by name

```terraform
data "emma_workflow_templates" "backup" {
  name_like = "backup"
}

output "backup_template_ids" {
  value = [for t in data.emma_workflow_templates.backup.workflow_templates : {
    id   = t.id
    name = t.name
  }]
}
```

## Schema

### Optional

- `name_like` (String) — Filter workflow templates by name using partial match.
- `status` (String) — Filter workflow templates by status: `PUBLISHED` or `UNPUBLISHED`.

### Read-Only

- `workflow_templates` (List of Object) — List of matching workflow templates. Each element has:
  - `id` (String) — ID of the workflow template.
  - `name` (String) — Name of the workflow template.
  - `description` (String) — Description of the workflow template.
  - `content_type` (String) — Format of the content field.
  - `content` (String) — Content of the workflow template.
  - `status` (String) — Status of the workflow template (`PUBLISHED` or `UNPUBLISHED`).
  - `resource_type` (String) — Type of the resource.
  - `created_at` (String) — Date and time when the workflow template was created (ISO 8601).
  - `created_by_name` (String) — Name of the user who created the workflow template.
  - `created_by_id` (String) — ID of the user who created the workflow template.
  - `modified_at` (String) — Date and time when the workflow template was last modified (ISO 8601).
  - `modified_by_name` (String) — Name of the user who last modified the workflow template.
  - `modified_by_id` (String) — ID of the user who last modified the workflow template.
  - `is_deleted` (Boolean) — Indicates whether the workflow template is deleted.
  - `is_content_valid` (Boolean) — Indicates whether the content of the workflow template is valid.
