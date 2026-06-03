---
page_title: "emma_spot_configurations Data Source - emma"
subcategory: ""
description: |-
  Returns available spot instance configurations filtered by GPU type, provider, data center, hardware specs, and price.
---

# emma_spot_configurations (Data Source)

Returns a list of available spot instance configurations. Use this data source to discover valid hardware configurations, pricing, and GPU availability before creating a spot instance.

All filter attributes are optional — omit them to get all configurations, or combine them to narrow results.

## Example Usage

### Find GPU spot configurations

```terraform
data "emma_accelerator_types" "all" {}

data "emma_spot_configurations" "gpu_t4" {
  accelerator_type_id = one([
    for t in data.emma_accelerator_types.all.accelerator_types :
    t.id if t.accelerator_type == "NVIDIA T4"
  ])
  vcpu_max = 4
}

output "t4_spot_configs" {
  value = [for c in data.emma_spot_configurations.gpu_t4.configurations : {
    data_center = c.data_center_id
    vcpu        = c.vcpu
    ram_gb      = c.ram_gb
    gpu         = c.accelerators
    price       = c.price_per_unit
  }]
}
```

### Find cheapest spot instances

```terraform
data "emma_spot_configurations" "cheap" {
  price_max  = 0.05
  vcpu_min   = 2
  ram_gb_min = 4
}

output "cheap_spots_count" {
  value = length(data.emma_spot_configurations.cheap.configurations)
}
```

## Schema

### Optional

- `accelerator_type_id` (String) Filter by GPU accelerator type ID. Use the `emma_accelerator_types` data source to find the ID.
- `data_center_id` (String) Filter by data center ID (e.g. "aws-sa-east-1", "azure-westeurope")
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

- `configurations` (List of Object) List of matching spot configurations. Each element has:
  - `provider_id` (Number) — Cloud provider ID
  - `provider_name` (String) — Cloud provider name (e.g. "AWS", "Azure")
  - `location_name` (String) — Location name
  - `data_center_id` (String) — Data center ID
  - `data_center_name` (String) — Data center name
  - `os_id` (Number) — Operating system ID
  - `os_type` (String) — Operating system type (e.g. "Ubuntu")
  - `cloud_network_types` (List of String) — Available network types
  - `vcpu_type` (String) — vCPU type
  - `vcpu` (Number) — Number of vCPUs
  - `ram_gb` (Number) — RAM in gigabytes
  - `volume_gb` (Number) — Volume size in gigabytes
  - `volume_type` (String) — Volume type
  - `accelerator_type_id` (String) — GPU accelerator type ID (empty if no GPU)
  - `accelerator_type` (String) — GPU model name (empty if no GPU)
  - `accelerators` (Number) — Number of GPU accelerators (0 if no GPU)
  - `price_per_unit` (Number) — Spot price per hour
  - `price_currency` (String) — Currency (e.g. "EUR")
  - `price_unit` (String) — Billing period (e.g. "HOURS")
