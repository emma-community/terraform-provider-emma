---
page_title: "emma_kuber_nodes_configs Data Source - emma"
subcategory: ""
description: |-
  Returns available hardware configurations for Kubernetes cluster worker nodes, filtered by connection type and optional hardware constraints.
---

# emma_kuber_nodes_configs (Data Source)

Returns a list of available hardware configurations for Kubernetes cluster worker nodes in the Emma platform. Use this data source to discover valid node types, CPU/RAM/storage options, GPU configurations, and pricing before creating or scaling a Kubernetes cluster.

`k8s_connection_type` is **required** — the API enforces it. All other filter attributes are optional.

## Example Usage

### List all public node configurations

```terraform
data "emma_kuber_nodes_configs" "public" {
  k8s_connection_type = "public"
}

output "node_config_count" {
  value = length(data.emma_kuber_nodes_configs.public.configurations)
}
```

### Find GPU-enabled node configs for a private cluster

```terraform
data "emma_kuber_nodes_configs" "gpu_private" {
  k8s_connection_type = "private"
  accelerator_type_id = "aaf51e31-b8d2-42dd-af72-cf3b6ec49370"
}

output "gpu_node_options" {
  value = [for c in data.emma_kuber_nodes_configs.gpu_private.configurations : {
    provider    = c.provider_name
    data_center = c.data_center_id
    vcpu        = c.vcpu
    ram_gb      = c.ram_gb
    gpu_count   = c.accelerators
    gpu_type    = c.accelerator_type
    price       = c.price_per_unit
  }]
}
```

### Find budget node configs under a price cap

```terraform
data "emma_kuber_nodes_configs" "cheap" {
  k8s_connection_type = "public"
  vcpu_min            = 4
  ram_gb_min          = 8
  price_max           = 100
}

output "affordable_nodes" {
  value = [for c in data.emma_kuber_nodes_configs.cheap.configurations : {
    id       = c.id
    provider = c.provider_name
    vcpu     = c.vcpu
    ram_gb   = c.ram_gb
    price    = c.price_per_unit
  }]
}
```

## Schema

### Required

- `k8s_connection_type` (String) — Kubernetes cluster network connectivity type. The API requires this value. Use `"public"` or `"private"`.

### Optional

- `provider_id` (Number) — Filter by cloud provider ID.
- `location_id` (Number) — Filter by geographic location ID.
- `accelerator_type_id` (String) — Filter by GPU accelerator type ID. Use `emma_accelerator_types` to discover available IDs.
- `accelerators_min` (Number) — Minimum quantity of GPU accelerators.
- `accelerators_max` (Number) — Maximum quantity of GPU accelerators.
- `data_center_id` (String) — Filter by data center ID (e.g. "aws-us-east-1").
- `vcpu_type` (String) — Filter by vCPU type: "shared", "standard", or "hpc".
- `vcpu_min` (Number) — Minimum number of vCPUs.
- `vcpu_max` (Number) — Maximum number of vCPUs.
- `ram_gb_min` (Number) — Minimum RAM in gigabytes.
- `ram_gb_max` (Number) — Maximum RAM in gigabytes.
- `volume_gb_min` (Number) — Minimum volume size in gigabytes.
- `volume_gb_max` (Number) — Maximum volume size in gigabytes.
- `volume_type` (String) — Filter by volume type (e.g. "ssd", "hdd").
- `price_min` (Number) — Minimum price per unit.
- `price_max` (Number) — Maximum price per unit.

### Read-Only

- `configurations` (List of Object) — List of matching Kubernetes node configurations. Each element has:
  - `id` (String) — Configuration ID
  - `provider_id` (Number) — ID of the cloud provider
  - `provider_name` (String) — Name of the cloud provider (e.g. "AWS", "Azure")
  - `location_id` (Number) — Location ID
  - `location_name` (String) — Location name (city or region)
  - `data_center_id` (String) — ID of the data center
  - `data_center_name` (String) — Name of the data center
  - `os_id` (Number) — ID of the operating system
  - `os_type` (String) — Type of the operating system (e.g. "Ubuntu")
  - `os_version` (String) — Version of the operating system (e.g. "22.04")
  - `cloud_network_types` (List of String) — Supported cloud network types (e.g. ["default", "isolated"])
  - `vcpu_type` (String) — vCPU type: "shared", "standard", or "hpc"
  - `vcpu` (Number) — Number of virtual CPUs
  - `ram_gb` (Number) — RAM in gigabytes
  - `volume_gb` (Number) — Volume size in gigabytes
  - `volume_type` (String) — Volume type (e.g. "ssd", "hdd")
  - `accelerator_type_id` (String) — GPU accelerator type ID. Empty when no GPU is available.
  - `accelerator_type` (String) — GPU accelerator type name (e.g. "NVIDIA T4"). Empty when no GPU is available.
  - `accelerators` (Number) — Quantity of GPU accelerators. 0 when no GPU is available.
  - `price_per_unit` (Number) — Price per billing unit for this configuration
  - `price_currency` (String) — Currency of the price (e.g. "EUR")
  - `price_unit` (String) — Billing period (e.g. "MONTHS")
