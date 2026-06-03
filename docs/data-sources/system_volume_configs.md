---
page_title: "emma_system_volume_configs Data Source - emma"
subcategory: ""
description: |-
  Returns available configurations for system volumes, with optional filtering by attachment target, minimum size, and volume type.
---

# emma_system_volume_configs (Data Source)

Returns a list of available system volume configurations in the Emma platform. Use this data source to discover valid sizes, types, and pricing for system volumes — for example, when planning a VM deployment or preparing to upsize an existing system disk.

All filter attributes are optional — omit them to get all available configurations, or combine them to narrow results.

Note: this data source does not support filtering by `provider_id`. Use the `provider_id` field in the returned configurations to filter results in your Terraform expressions.

## Example Usage

### List all system volume configurations

```terraform
data "emma_system_volume_configs" "all" {}

output "config_count" {
  value = length(data.emma_system_volume_configs.all.configurations)
}
```

### Find SSD configurations of at least 100 GB

```terraform
data "emma_system_volume_configs" "large_ssd" {
  volume_gb_min = 100
  volume_type   = "ssd"
}

output "large_ssd_options" {
  value = [for c in data.emma_system_volume_configs.large_ssd.configurations : {
    data_center = c.data_center_id
    size_gb     = c.volume_gb
    price       = c.price_per_unit
    currency    = c.price_currency
  }]
}
```

### Find valid upsizing options for an attached VM's system volume

```terraform
data "emma_system_volume_configs" "vm_options" {
  attached_to_id = 4521
  volume_gb_min  = 50
}

output "upsize_options" {
  value = data.emma_system_volume_configs.vm_options.configurations
}
```

## Schema

### Optional

- `attached_to_id` (Number) — ID of the compute instance the system volume is attached to. Use to scope results to valid configurations for a specific VM.
- `volume_gb_min` (Number) — Minimum volume size in gigabytes to filter configurations.
- `volume_type` (String) — Volume type to filter configurations (e.g. "ssd", "hdd").

### Read-Only

- `configurations` (List of Object) — List of matching system volume configurations. Each element has:
  - `provider_id` (Number) — ID of the cloud provider
  - `provider_name` (String) — Name of the cloud provider (e.g. "AWS", "Azure")
  - `location_id` (Number) — Location ID
  - `location_name` (String) — Location name (city or region)
  - `data_center_id` (String) — ID of the data center (e.g. "aws-us-east-1")
  - `data_center_name` (String) — Name of the data center
  - `volume_gb` (Number) — Volume size in gigabytes
  - `volume_type` (String) — Volume type (e.g. "ssd", "hdd")
  - `price_per_unit` (Number) — Price per billing unit for this configuration
  - `price_currency` (String) — Currency of the price (e.g. "EUR")
  - `price_unit` (String) — Billing period (e.g. "MONTHS")
