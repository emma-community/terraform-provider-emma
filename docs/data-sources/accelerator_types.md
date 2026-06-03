---
page_title: "emma_accelerator_types Data Source - emma"
subcategory: ""
description: |-
  Returns a list of all available GPU accelerator types that can be used with compute instances.
---

# emma_accelerator_types (Data Source)

Returns a list of all available GPU accelerator types. Use this data source to discover which GPU models are available on the platform before creating GPU-enabled virtual machines or spot instances, and to resolve a model name to its `accelerator_type_id`.

## Example Usage

### Resolve a GPU id by name

```terraform
data "emma_accelerator_types" "all" {}

resource "emma_vm" "gpu_vm" {
  # ...other required fields...
  accelerator_type_id = data.emma_accelerator_types.all.ids_by_name["NVIDIA T4"]
  accelerators        = 1
}
```

### Use with vm_configurations to find GPU VMs

```terraform
data "emma_accelerator_types" "all" {}

data "emma_vm_configurations" "gpu" {
  accelerator_type_id = data.emma_accelerator_types.all.ids_by_name["NVIDIA A100 40 GB"]
}
```

### List the names of all available GPUs

Handy for debugging or for CI checks:

```terraform
data "emma_accelerator_types" "all" {}

output "available_gpus" {
  value = keys(data.emma_accelerator_types.all.ids_by_name)
}
```

## Schema

### Read-Only

- `accelerator_types` (List of Object) Full list of available GPU accelerator types. Each element has:
  - `id` (String) — Unique ID of the accelerator type (UUID)
  - `accelerator_type` (String) — Name of the GPU model (e.g. `"NVIDIA T4"`, `"AMD Instinct MI25"`)
- `ids_by_name` (Map of String) Map from accelerator type name to its id, e.g. `accelerator_type_id = data.emma_accelerator_types.all.ids_by_name["NVIDIA T4"]`. Lookup of a name that is not in the catalogue returns `null`.
