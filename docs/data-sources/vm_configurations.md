---
page_title: "emma_vm_configurations Data Source - emma"
subcategory: ""
description: |-
  Returns available virtual machine configurations filtered by GPU type, provider, data center, hardware specs, and price.
---

# emma_vm_configurations (Data Source)

Returns a list of available virtual machine configurations. Use this data source to discover valid hardware configurations, pricing, and GPU availability before creating a virtual machine.

All filter attributes are optional — omit them to get all configurations, or combine them to narrow results.

## Example Usage

### Find cheapest GPU VM configurations

```terraform
data "emma_accelerator_types" "all" {}

data "emma_vm_configurations" "gpu_t4" {
  accelerator_type_id = data.emma_accelerator_types.all.ids_by_name["NVIDIA T4"]
  price_max           = 400
}

output "t4_configs" {
  value = [for c in data.emma_vm_configurations.gpu_t4.configurations : {
    provider    = c.provider_name
    data_center = c.data_center_id
    vcpu        = c.vcpu
    ram_gb      = c.ram_gb
    gpu         = c.accelerators
    price       = c.price_per_unit
  }]
}
```

### Filter by provider and data center

```terraform
data "emma_vm_configurations" "azure_gpu" {
  accelerator_type_id = "aaf51e31-b8d2-42dd-af72-cf3b6ec49370"
  provider_id         = 256
  ram_gb_min          = 14
}

output "azure_gpu_configs" {
  value = data.emma_vm_configurations.azure_gpu.configurations
}
```

### Find high-memory VMs

```terraform
data "emma_vm_configurations" "high_mem" {
  ram_gb_min = 64
  vcpu_min   = 8
  vcpu_type  = "standard"
}

output "high_mem_configs_count" {
  value = length(data.emma_vm_configurations.high_mem.configurations)
}
```

## Schema

### Optional

- `accelerator_type_id` (String) Filter by GPU accelerator type ID. Use the `emma_accelerator_types` data source to find the ID.
- `data_center_id` (String) Filter by data center ID (e.g. "aws-us-east-2", "azure-westeurope")
- `provider_id` (Number) Filter by cloud provider ID (e.g. 255 = AWS, 256 = Azure)
- `vcpu_type` (String) Filter by vCPU type: "shared", "standard", or "hpc"
- `vcpu_min` (Number) Minimum number of vCPUs
- `vcpu_max` (Number) Maximum number of vCPUs
- `ram_gb_min` (Number) Minimum RAM in gigabytes
- `ram_gb_max` (Number) Maximum RAM in gigabytes
- `volume_gb_min` (Number) Minimum volume size in gigabytes
- `volume_gb_max` (Number) Maximum volume size in gigabytes
- `price_min` (Number) Minimum price per unit
- `price_max` (Number) Maximum price per unit

### Read-Only

- `configurations` (List of Object) List of matching VM configurations. Each element has:
  - `provider_id` (Number) — Cloud provider ID
  - `provider_name` (String) — Cloud provider name (e.g. "AWS", "Azure")
  - `location_name` (String) — Location name
  - `data_center_id` (String) — Data center ID
  - `data_center_name` (String) — Data center name
  - `os_id` (Number) — Operating system ID
  - `os_type` (String) — Operating system type (e.g. "Ubuntu")
  - `cloud_network_types` (List of String) — Available network types (e.g. ["default", "isolated", "multi-cloud"])
  - `vcpu_type` (String) — vCPU type
  - `vcpu` (Number) — Number of vCPUs
  - `ram_gb` (Number) — RAM in gigabytes
  - `volume_gb` (Number) — Volume size in gigabytes
  - `volume_type` (String) — Volume type (e.g. "ssd", "ssd-plus")
  - `accelerator_type_id` (String) — GPU accelerator type ID (empty if no GPU)
  - `accelerator_type` (String) — GPU model name (empty if no GPU)
  - `accelerators` (Number) — Number of GPU accelerators (0 if no GPU)
  - `price_per_unit` (Number) — Price per unit period
  - `price_currency` (String) — Currency (e.g. "EUR")
  - `price_unit` (String) — Billing period (e.g. "MONTHS")
