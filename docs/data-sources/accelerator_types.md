---
page_title: "emma_accelerator_types Data Source - emma"
subcategory: ""
description: |-
  Returns a list of all available GPU accelerator types that can be used with compute instances.
---

# emma_accelerator_types (Data Source)

Returns a list of all available GPU accelerator types. Use this data source to discover which GPU models are available on the platform before creating GPU-enabled virtual machines or spot instances, and to resolve a model name to its `accelerator_type_id`.

## Example Usage

### List all available GPU types

```terraform
data "emma_accelerator_types" "all" {}

output "available_gpus" {
  value = data.emma_accelerator_types.all.accelerator_types
}
```

### Use with vm_configurations to find GPU VMs

```terraform
data "emma_accelerator_types" "all" {}

# Pick the first accelerator type and look up VM configs for it
data "emma_vm_configurations" "gpu" {
  accelerator_type_id = data.emma_accelerator_types.all.accelerator_types[0].id
}

output "gpu_types" {
  value = [for t in data.emma_accelerator_types.all.accelerator_types : t.accelerator_type]
}
```

## Schema

### Read-Only

- `accelerator_types` (List of Object) List of available GPU accelerator types. Each element has:
  - `id` (String) — Unique ID of the accelerator type (UUID)
  - `accelerator_type` (String) — Name of the GPU model (e.g. "NVIDIA T4", "AMD Instinct MI25")
