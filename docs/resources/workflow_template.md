---
page_title: "emma_workflow_template Resource - emma"
subcategory: ""
description: |-
  Manages a workflow template. A workflow template is a predefined blueprint for a workflow that includes shell commands and parameters. It serves as a starting point for creating workflows, allowing users to quickly set up and execute common processes.
---

# emma_workflow_template (Resource)

Manages a workflow template in Emma. A workflow template is a predefined blueprint for a workflow that includes shell commands and parameters. It serves as a starting point for creating workflows, allowing users to quickly set up and execute common processes.

To list existing templates without managing them, use the [`emma_workflow_templates` data source](../data-sources/workflow_templates.md).

## Example Usage

### Minimal template

```terraform
resource "emma_workflow_template" "backup" {
  name          = "daily-backup"
  content_type  = "Shell"
  status        = "PUBLISHED"
  resource_type = "COMPUTE"

  content = <<-EOT
    #!/bin/bash
    set -e
    echo "Starting backup at $(date)"
    tar czf /tmp/backup.tar.gz /var/data
    echo "Backup complete"
  EOT
}
```

### Template with resource constraints, content parameters, and tags

```terraform
resource "emma_workflow_template" "deploy" {
  name          = "app-deploy"
  description   = "Deploys the application to a target VM"
  content_type  = "Shell"
  status        = "PUBLISHED"
  resource_type = "COMPUTE"

  content = <<-EOT
    #!/bin/bash
    set -e
    APP_VERSION="${APP_VERSION}"
    TARGET_ENV="${TARGET_ENV}"
    echo "Deploying $APP_VERSION to $TARGET_ENV"
    # deployment steps here
  EOT

  resource_params = [
    {
      name  = "minCpuCores"
      value = "4"
    },
    {
      name  = "minRamSizeGb"
      value = "8"
    },
  ]

  content_params = [
    {
      name          = "APP_VERSION"
      default_value = "latest"
      mandatory     = true
    },
    {
      name          = "TARGET_ENV"
      default_value = "production"
      mandatory     = false
    },
  ]

  tags = [
    {
      tag_id = null
      key    = "team"
      value  = "platform"
    },
    {
      tag_id = null
      key    = "environment"
      value  = "production"
    },
  ]
}

output "template_id" {
  value = emma_workflow_template.deploy.id
}
```

## Schema

### Required

- `name` (String) — Name of the workflow template.
- `content_type` (String) — Format of the content field. Only `Shell` is supported.
- `content` (String) — Content of the workflow template. For `Shell` content type, contains the shell commands to execute.
- `status` (String) — Status of the workflow template. Valid values: `PUBLISHED`, `UNPUBLISHED`.
- `resource_type` (String) — Type of the resource. Use `COMPUTE`.

### Optional

- `description` (String) — Description of the workflow template.
- `resource_params` (List of Object) — Parameters that constrain the resources a workflow may run on (e.g. minimum CPU, RAM). Each object has:
  - `name` (String) — Parameter name (e.g. `minCpuCores`, `minRamSizeGb`, `minVolumeSizeGb`).
  - `value` (String) — Parameter value.
- `content_params` (List of Object) — Template placeholder parameters. These fill in variables in `content` when a workflow is created from this template. Each object has:
  - `name` (String) — Name of the content parameter.
  - `default_value` (String) — Default value of the parameter.
  - `mandatory` (Boolean) — Whether the parameter must be supplied when creating a workflow.
- `tags` (List of Object) — Tags assigned to the workflow template. Each object has:
  - `tag_id` (String) — ID of an existing tag (may be `null` when creating new tags).
  - `key` (String) — Tag key (e.g. `environment`, `team`).
  - `value` (String) — Tag value.

### Read-Only

- `id` (String) — ID of the workflow template.
- `created_at` (String) — Date and time when the workflow template was created (ISO 8601).
- `created_by_name` (String) — Name of the user who created the workflow template.
- `created_by_id` (String) — ID of the user who created the workflow template.
- `modified_at` (String) — Date and time when the workflow template was last modified (ISO 8601).
- `modified_by_name` (String) — Name of the user who last modified the workflow template.
- `modified_by_id` (String) — ID of the user who last modified the workflow template.
- `is_deleted` (Boolean) — Indicates whether the workflow template is deleted.
- `is_content_valid` (Boolean) — Indicates whether the content of the workflow template is valid.

## Import

Workflow templates can be imported using their ID:

```shell
terraform import emma_workflow_template.foo <id>
```
