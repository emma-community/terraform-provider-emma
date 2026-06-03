---
page_title: "emma_vm_resize_configs Data Source - emma"
subcategory: ""
description: |-
  Returns available hardware configurations for resizing an existing virtual machine, with optional filtering by VM ID and hardware constraints.
---

# emma_vm_resize_configs (Data Source)

Returns a list of valid resize configurations for an existing virtual machine. Use this data source to discover which vCPU, RAM, and GPU combinations are available for a given VM before calling the VM edit endpoint.

Unlike `emma_vm_configurations`, this data source does not return provider, location, or data center fields — results are scoped to configurations compatible with the VM's existing placement. The configurations include the provider-specific compute type identifier (`provider_compute_type`) needed for the resize API call.

All filter attributes are optional.

## Example Usage

### List all resize configs for a specific VM

```terraform
data "emma_vm_resize_configs" "vm_4521" {
  vm_id = 4521
}

output "resize_options" {
  value = [for c in data.emma_vm_resize_configs.vm_4521.configurations : {
    vcpu     = c.vcpu
    ram_gb   = c.ram_gb
    price    = c.price_per_unit
    currency = c.price_currency
  }]
}
```

### Find resize configs with at least 16 vCPUs and 64 GB RAM under a price cap

```terraform
data "emma_vm_resize_configs" "upgrade" {
  vm_id     = 4521
  vcpu_min  = 16
  ram_gb_min = 64
  price_max = 500
}

output "upgrade_candidates" {
  value = data.emma_vm_resize_configs.upgrade.configurations
}
```

### Find GPU resize configs for a VM

```terraform
data "emma_vm_resize_configs" "gpu_upgrade" {
  vm_id = 4521
}

output "gpu_resize_options" {
  value = [
    for c in data.emma_vm_resize_configs.gpu_upgrade.configurations : c
    if c.accelerator_type != null && c.accelerator_type != ""
  ]
}
```

## Schema

### Optional

- `vm_id` (Number) — ID of the virtual machine to retrieve resize configurations for. When provided, results are scoped to configurations compatible with the VM's current placement.
- `vcpu_type` (String) — Filter by vCPU type: "shared", "standard", or "hpc".
- `vcpu_min` (Number) — Minimum number of vCPUs.
- `vcpu_max` (Number) — Maximum number of vCPUs.
- `ram_gb_min` (Number) — Minimum RAM in gigabytes.
- `ram_gb_max` (Number) — Maximum RAM in gigabytes.
- `price_min` (Number) — Minimum price per unit.
- `price_max` (Number) — Maximum price per unit.

### Read-Only

- `configurations` (List of Object) — List of available VM resize configurations. Each element has:
  - `vcpu_type` (String) — vCPU type: "shared", "standard", or "hpc"
  - `vcpu` (Number) — Number of virtual CPUs
  - `ram_gb` (Number) — RAM in gigabytes
  - `provider_compute_type` (String) — Provider-specific compute type identifier (used in the resize API call)
  - `accelerator_type_id` (String) — GPU accelerator type ID. Empty when no GPU is available.
  - `accelerator_type` (String) — GPU accelerator type name (e.g. "NVIDIA T4"). Empty when no GPU is available.
  - `accelerators` (Number) — Quantity of GPU accelerators. 0 when no GPU is available.
  - `price_per_unit` (Number) — Price per billing unit for this configuration
  - `price_currency` (String) — Currency of the price (e.g. "EUR")
  - `price_unit` (String) — Billing period (e.g. "MONTHS")
