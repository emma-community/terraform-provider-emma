---
page_title: "emma_ssh_keys Data Source - emma"
subcategory: ""
description: |-
  Returns a list of all SSH keys in the project.
---

# emma_ssh_keys (Data Source)

Returns a list of all SSH keys stored in the Emma platform. Use this data source to look up SSH key IDs for assigning to compute instances, or to verify which keys are available before creating a VM.

The full public key content is included in the `key` field, which is useful for cross-referencing keys against your local `~/.ssh` directory or for auditing team access.

This data source returns all SSH keys accessible to your credentials and does not accept filter arguments.

## Example Usage

### List all SSH key names and fingerprints

```terraform
data "emma_ssh_keys" "all" {}

output "key_inventory" {
  value = [for k in data.emma_ssh_keys.all.ssh_keys : {
    name        = k.name
    fingerprint = k.fingerprint
    type        = k.key_type
  }]
}
```

### Resolve an SSH key ID by name for use in a VM resource

```terraform
data "emma_ssh_keys" "all" {}

locals {
  deploy_key_id = one([
    for k in data.emma_ssh_keys.all.ssh_keys : k.id
    if k.name == "ci-deploy"
  ])
}

resource "emma_vm" "web" {
  # ...
  ssh_key_id = local.deploy_key_id
}
```

### Find all ED25519 keys created by a specific user

```terraform
data "emma_ssh_keys" "all" {}

output "ed25519_keys" {
  value = [
    for k in data.emma_ssh_keys.all.ssh_keys : k
    if k.key_type == "ED25519" && k.created_by_name == "a.prihodko"
  ]
}
```

## Schema

### Read-Only

- `ssh_keys` (List of Object) — List of all accessible SSH keys. Each element has:
  - `id` (Number) — SSH key ID
  - `name` (String) — SSH key name
  - `key` (String) — Full SSH public key content (e.g. "ssh-rsa AAAA...")
  - `key_type` (String) — SSH key algorithm: "RSA" or "ED25519"
  - `fingerprint` (String) — SHA-256 fingerprint of the public key
  - `created_at` (String) — Date and time when the SSH key was created (ISO 8601)
  - `created_by_name` (String) — Name of the user who created the SSH key
  - `created_by_id` (Number) — ID of the user who created the SSH key
