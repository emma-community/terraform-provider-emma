---
page_title: "emma_ssh_key Data Source - emma"
subcategory: ""
description: |-
  Provides information about an existing SSH key by its ID.
---

# emma_ssh_key (Data Source)

Provides information about an existing SSH key by its ID.

Use this data source to look up an SSH key that already exists in Emma — for example, to wire its `id` into an `emma_vm` resource without importing the key into Terraform state.

To create a new SSH key and manage it as Terraform-owned infrastructure, use the [`emma_ssh_key` resource](../resources/ssh_key.md) instead.

## Example Usage

```terraform
data "emma_ssh_key" "shared" {
  id = 42
}

resource "emma_vm" "web" {
  name               = "web-server"
  data_center_id     = data.emma_data_center.aws.id
  os_id              = data.emma_operating_system.ubuntu.id
  cloud_network_type = "multi-cloud"
  vcpu_type          = "shared"
  vcpu               = 2
  ram_gb             = 4
  volume_type        = "ssd"
  volume_gb          = 40
  ssh_key_id         = data.emma_ssh_key.shared.id
}
```

## Schema

### Required

- `id` (Number) — ID of the SSH key to look up.

### Read-Only

- `name` (String) — SSH key name.
- `key` (String) — SSH public key content.
- `key_type` (String) — SSH key type: `RSA` or `ED25519`.
- `fingerprint` (String) — SSH key fingerprint.
- `created_at` (String) — Date and time when the SSH key was created (ISO 8601).
- `created_by_name` (String) — Name of the user who created the SSH key.
- `created_by_id` (Number) — ID of the user who created the SSH key.
