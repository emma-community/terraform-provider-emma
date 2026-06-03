---
page_title: "emma_statistical_data Data Source - emma"
subcategory: ""
description: |-
  Fetches statistical data from emma's DWH. Exactly one query block must be configured. The corresponding result list attribute will be populated in state.
---

# emma_statistical_data (Data Source)

!> **Deprecated** The underlying API endpoint `POST /v1/statistics/retrieve` was marked deprecated on **2026-05-13**. This data source remains available for historical reporting until a replacement endpoint is provided. Do not build new integrations against it.

Fetches statistical data from Emma's data warehouse. The data source implements a discriminated-union query model: configure **exactly one** of the ten query blocks below. The provider will return a validation error if zero or more than one block is set.

Only the result list attribute that corresponds to the configured query block will be populated — all other `*_results` attributes will be empty lists.

## Query Variants

| Query block | Result attribute | Use case |
|---|---|---|
| `project_summary` | `project_summary_results` | Cost + installation count per service for the whole project |
| `resource_analysis` | `resource_analysis_results` | CPU / RAM / disk time-series over a custom date range |
| `vm_analytics` | `vm_analytics_results` | Aggregated utilization analytics for a single VM |
| `vm_monitoring` | `vm_monitoring_results` | Time-series monitoring data for a single VM |
| `product_statistics` | `product_statistics_results` | Per-VM cost and resource statistics, filterable by service |
| `kubernetes_cluster_objects` | `kubernetes_cluster_objects_results` | Inventory of Kubernetes objects in a cluster |
| `kubernetes_cluster_metrics` | `kubernetes_cluster_metrics_results` | Available metric definitions for a cluster |
| `kubernetes_cluster_object_states` | `kubernetes_cluster_object_states_results` | Object state distributions in a cluster |
| `kubernetes_cluster_current_state` | `kubernetes_cluster_current_state_results` | Current-state metric values for cluster objects |
| `kubernetes_cluster_changing_metrics` | `kubernetes_cluster_changing_metrics_results` | Time-series changing metrics for cluster objects |

## Example Usage

### Project cost summary

```terraform
data "emma_statistical_data" "summary" {
  project_summary = {
    dataset_name = "project_summary"
  }
}

output "service_costs" {
  value = [for r in data.emma_statistical_data.summary.project_summary_results : {
    service           = r.service
    cost              = r.cost
    all_installations = r.all_installations
  }]
}
```

### VM analytics for a specific VM

```terraform
data "emma_statistical_data" "vm_stats" {
  vm_analytics = {
    dataset_name = "vm_analytics"
    vm_id        = 12345
  }
}

output "cpu_utilization" {
  value = [for r in data.emma_statistical_data.vm_stats.vm_analytics_results : {
    timecode                       = r.timecode
    cpu_utilization_average_cores  = r.cpu_utilization_average_cores
    ram_usage_average_mb           = r.ram_usage_average_mb
  }]
}
```

### Resource analysis over a date range

```terraform
data "emma_statistical_data" "resource_trend" {
  resource_analysis = {
    dataset_name = "resource_analysis"
    filters = {
      period_start = "2026-01-01T00:00:00Z"
      period_end   = "2026-03-31T23:59:59Z"
    }
  }
}

output "resource_trend" {
  value = data.emma_statistical_data.resource_trend.resource_analysis_results
}
```

## Schema

### Query Input Blocks (Optional — exactly one must be set)

#### `project_summary`

- `dataset_name` (String, Required) — Dataset name identifier.

#### `resource_analysis`

- `dataset_name` (String, Required) — Dataset name identifier.
- `filters` (Required):
  - `period_start` (String, Required) — Start of the analysis period (ISO 8601).
  - `period_end` (String, Required) — End of the analysis period (ISO 8601).

#### `vm_analytics`

- `dataset_name` (String, Required) — Dataset name identifier.
- `vm_id` (Number, Required) — ID of the virtual machine.

#### `vm_monitoring`

- `dataset_name` (String, Required) — Dataset name identifier.
- `vm_id` (Number, Required) — ID of the virtual machine.
- `filters` (Required):
  - `period` (String, Required) — Monitoring period (e.g. `1h`, `24h`, `7d`).

#### `product_statistics`

- `dataset_name` (String, Required) — Dataset name identifier.
- `filters` (Required):
  - `service_filter` (String, Required) — Service name filter.

#### `kubernetes_cluster_objects`

- `dataset_name` (String, Required) — Dataset name identifier.
- `core_cluster_id` (Number, Required) — Core cluster ID.

#### `kubernetes_cluster_metrics`

- `dataset_name` (String, Required) — Dataset name identifier.
- `core_cluster_id` (Number, Required) — Core cluster ID.
- `filters` (Required):
  - `object_type` (String, Required) — Object type to filter.
  - `object_name` (String, Required) — Object name to filter.
  - `breakdown_level` (String, Required) — Breakdown level.

#### `kubernetes_cluster_object_states`

- `dataset_name` (String, Required) — Dataset name identifier.
- `core_cluster_id` (Number, Required) — Core cluster ID.
- `filters` (Required):
  - `object_type` (String, Required) — Object type to filter.
  - `object_name` (String, Required) — Object name to filter.
  - `breakdown_level` (String, Required) — Breakdown level.
  - `object_states_metrics` (List of String, Required) — List of object state metric names.

#### `kubernetes_cluster_current_state`

- `dataset_name` (String, Required) — Dataset name identifier.
- `core_cluster_id` (Number, Required) — Core cluster ID.
- `filters` (Required):
  - `object_type` (String, Required) — Object type to filter.
  - `object_name` (String, Required) — Object name to filter.
  - `breakdown_level` (String, Required) — Breakdown level.
  - `current_state_metrics` (List of String, Required) — List of current state metric names.
  - `custom_filter_state` (String, Required) — Custom filter state.
  - `custom_filter_avg_cpu_rule` (String, Optional) — Custom CPU filter rule (nullable).
  - `custom_filter_avg_cpu_value` (Number, Required) — Custom CPU filter value.
  - `custom_filter_avg_memory_rule` (String, Required) — Custom memory filter rule.
  - `custom_filter_avg_memory_value` (Number, Required) — Custom memory filter value.
  - `custom_filter_avg_storage_rule` (String, Required) — Custom storage filter rule.
  - `custom_filter_avg_storage_value` (Number, Required) — Custom storage filter value.
  - `custom_filter_subobjects` (List of String, Required) — List of subobjects to filter.

#### `kubernetes_cluster_changing_metrics`

- `dataset_name` (String, Required) — Dataset name identifier.
- `core_cluster_id` (Number, Required) — Core cluster ID.
- `filters` (Required):
  - `object_type` (String, Required) — Object type to filter.
  - `object_name` (String, Required) — Object name to filter.
  - `breakdown_level` (String, Required) — Breakdown level.
  - `changing_metrics` (List of String, Required) — List of changing metric names.
  - `timespan` (String, Required) — Timespan for the query.
  - `custom_filter_state` (String, Required) — Custom filter state.
  - `custom_filter_avg_cpu_rule` (String, Optional) — Custom CPU filter rule (nullable).
  - `custom_filter_avg_cpu_value` (Number, Required) — Custom CPU filter value.
  - `custom_filter_avg_memory_rule` (String, Required) — Custom memory filter rule.
  - `custom_filter_avg_memory_value` (Number, Required) — Custom memory filter value.
  - `custom_filter_avg_storage_rule` (String, Required) — Custom storage filter rule.
  - `custom_filter_avg_storage_value` (Number, Required) — Custom storage filter value.
  - `custom_filter_subobjects` (List of String, Required) — List of subobjects to filter.

---

### Read-Only Result Attributes

Only the result list matching the configured query block will be non-empty.

#### `project_summary_results` (List of Object)

- `service` (String) — Service name.
- `all_installations` (Number) — Total number of installations.
- `cost` (Number) — Total cost.

#### `resource_analysis_results` (List of Object)

- `service` (String) — Service name.
- `timecode` (String) — Time code.
- `cpu_cores_number` (Number) — Number of CPU cores.
- `ram_total_amount_gb` (Number) — Total RAM in GB.
- `disk_space_total_gb` (Number) — Total disk space in GB.
- `type` (String) — Resource type.

#### `vm_analytics_results` (List of Object)

- `vm_id` (Number) — VM ID.
- `timecode` (String) — Time code.
- `avg_date_start` (String) — Average period start date.
- `avg_date_end` (String) — Average period end date.
- `quantiles_date_start` (String) — Quantiles period start date.
- `quantiles_date_end` (String) — Quantiles period end date.
- `cpu_data_present` (Number) — CPU data present flag.
- `cpu_utilization_average_cores` (Number) — Average CPU utilization in cores.
- `cpu_cores_number` (Number) — Total CPU cores.
- `cpu_total_percent` (Number) — CPU total percent.
- `cpu_human_label` (String) — Human-readable CPU label.
- `ram_data_present` (Number) — RAM data present flag.
- `ram_usage_average_mb` (Number) — Average RAM usage in MB.
- `ram_total_amount_mb` (Number) — Total RAM in MB.
- `ram_human_label` (String) — Human-readable RAM label.
- `disk_used_data_present` (Number) — Disk used data present flag.
- `disk_space_used_gb` (Number) — Disk space used in GB.
- `disk_space_total_gb` (Number) — Total disk space in GB.
- `disk_space_human_label` (String) — Human-readable disk label.
- `disk_write_data_present` (Number) — Disk write data present flag.
- `disk_write_bps` (Number) — Disk write speed in bytes/s.
- `disk_write_human` (Number) — Disk write human value.
- `disk_write_human_label` (String) — Human-readable disk write label.
- `disk_read_data_present` (Number) — Disk read data present flag.
- `disk_read_bps` (Number) — Disk read speed in bytes/s.
- `disk_read_human` (Number) — Disk read human value.
- `disk_read_human_label` (String) — Human-readable disk read label.
- `network_out_data_present` (Number) — Network out data present flag.
- `network_out_bps` (Number) — Network out speed in bytes/s.
- `network_out_human` (Number) — Network out human value.
- `network_out_human_label` (String) — Human-readable network out label.
- `network_in_data_present` (Number) — Network in data present flag.
- `network_in_bps` (Number) — Network in speed in bytes/s.
- `network_in_human` (Number) — Network in human value.
- `network_in_human_label` (String) — Human-readable network in label.
- `gpu_data_present` (Number) — GPU data present flag.
- `gpu_utilization_avg_percent` (Number) — Average GPU utilization percent.
- `gpu_total_percent` (Number) — GPU total percent.
- `gpu_human_label` (String) — Human-readable GPU label.
- `gpu_ram_data_present` (Number) — GPU RAM data present flag.
- `gpu_ram_usage_avg_gb` (Number) — Average GPU RAM usage in GB.
- `gpu_ram_total_gb` (Number) — Total GPU RAM in GB.
- `gpu_ram_human_label` (String) — Human-readable GPU RAM label.
- `gpu_ram_utilization_data_present` (Number) — GPU RAM utilization data present flag.
- `gpu_ram_utilization_avg_percent` (Number) — Average GPU RAM utilization percent.
- `gpu_ram_utilization_total_percent` (Number) — GPU RAM utilization total percent.
- `gpu_ram_utilization_human_label` (String) — Human-readable GPU RAM utilization label.
- `is_shown_short` (Number) — Is shown short flag.
- `type` (String) — Record type.

#### `vm_monitoring_results` (List of Object)

- `timecode` (String) — Time code.
- `cpu_data_present` (Number) — CPU data present flag.
- `cpu_utilization_average_cores` (Number) — Average CPU utilization in cores.
- `avg_cpu_utilization_average_cores` (Number) — Period average CPU utilization in cores.
- `max_cpu_utilization_average_cores` (Number) — Period max CPU utilization in cores.
- `cpu_total_percent` (Number) — CPU total percent.
- `cpu_human_label` (String) — Human-readable CPU label.
- `ram_data_present` (Number) — RAM data present flag.
- `ram_usage_average_gb` (Number) — Average RAM usage in GB.
- `avg_ram_usage_average_gb` (Number) — Period average RAM usage in GB.
- `max_ram_usage_average_gb` (Number) — Period max RAM usage in GB.
- `ram_total_amount_gb` (Number) — Total RAM in GB.
- `ram_human_label` (String) — Human-readable RAM label.
- `disk_used_data_present` (Number) — Disk used data present flag.
- `disk_space_used_gb` (Number) — Disk space used in GB.
- `avg_disk_space_used_gb` (Number) — Period average disk space used in GB.
- `max_disk_space_used_gb` (Number) — Period max disk space used in GB.
- `disk_space_total_gb` (Number) — Total disk space in GB.
- `disk_space_human_label` (String) — Human-readable disk label.
- `disk_write_data_present` (Number) — Disk write data present flag.
- `disk_write_coef` (Number) — Disk write coefficient.
- `disk_write_human` (Number) — Disk write human value.
- `avg_disk_write_human` (Number) — Period average disk write human value.
- `max_disk_write_human` (Number) — Period max disk write human value.
- `disk_write_human_label` (String) — Human-readable disk write label.
- `disk_read_data_present` (Number) — Disk read data present flag.
- `disk_read_coef` (Number) — Disk read coefficient.
- `disk_read_human` (Number) — Disk read human value.
- `avg_disk_read_human` (Number) — Period average disk read human value.
- `max_disk_read_human` (Number) — Period max disk read human value.
- `disk_read_human_label` (String) — Human-readable disk read label.
- `network_out_data_present` (Number) — Network out data present flag.
- `network_out_coef` (Number) — Network out coefficient.
- `network_out_human` (Number) — Network out human value.
- `avg_network_out_human` (Number) — Period average network out human value.
- `max_network_out_human` (Number) — Period max network out human value.
- `network_out_human_label` (String) — Human-readable network out label.
- `network_in_data_present` (Number) — Network in data present flag.
- `network_in_coef` (Number) — Network in coefficient.
- `network_in_human` (Number) — Network in human value.
- `avg_network_in_human` (Number) — Period average network in human value.
- `max_network_in_human` (Number) — Period max network in human value.
- `network_in_human_label` (String) — Human-readable network in label.
- `gpu_ram_usage_data_present` (Number) — GPU RAM usage data present flag.
- `gpu_ram_usage_avg_gb` (Number) — Average GPU RAM usage in GB.
- `avg_gpu_ram_usage_avg_gb` (Number) — Period average GPU RAM usage in GB.
- `max_gpu_ram_usage_avg_gb` (Number) — Period max GPU RAM usage in GB.
- `gpu_ram_usage_human_label` (String) — Human-readable GPU RAM usage label.
- `gpu_ram_utilization_avg_present` (Number) — GPU RAM utilization avg present flag.
- `gpu_ram_utilization_avg_percent` (Number) — Average GPU RAM utilization percent.
- `avg_gpu_ram_utilization_avg_percent` (Number) — Period average GPU RAM utilization percent.
- `max_gpu_ram_utilization_avg_percent` (Number) — Period max GPU RAM utilization percent.
- `gpu_ram_utilization_human_label` (String) — Human-readable GPU RAM utilization label.
- `gpu_utilization_data_present` (Number) — GPU utilization data present flag.
- `gpu_utilization_avg_percent` (Number) — Average GPU utilization percent.
- `avg_gpu_utilization_avg_percent` (Number) — Period average GPU utilization percent.
- `max_gpu_utilization_avg_percent` (Number) — Period max GPU utilization percent.
- `gpu_utilization_human_label` (String) — Human-readable GPU utilization label.

#### `product_statistics_results` (List of Object)

- `service` (String) — Service name.
- `vm_id` (Number) — VM ID.
- `vm_name` (String) — VM name.
- `head_product_id` (Number) — Head product ID.
- `head_product_name` (String) — Head product name.
- `currency` (String) — Cost currency.
- `cost` (Number) — Cost value.
- `provider_name` (String) — Cloud provider name.
- `country` (String) — Country.
- `location` (String) — Location.
- `latitude` (Number) — Geographic latitude.
- `longitude` (Number) — Geographic longitude.
- `status_normalized` (String) — Normalized status.
- `cpu_total` (Number) — Total CPU.
- `ram_total` (Number) — Total RAM.
- `disk_usage_total` (Number) — Total disk usage.
- `cpu_usage` (Number) — CPU usage.
- `ram_usage` (Number) — RAM usage.
- `disk_usage` (Number) — Disk usage.
- `empty_value` (Number) — Empty value placeholder.

#### `kubernetes_cluster_objects_results` (List of Object)

- `object_type` (String) — Type of the Kubernetes object.
- `object_name` (String) — Name of the Kubernetes object.

#### `kubernetes_cluster_metrics_results` (List of Object)

- `metric_name` (String) — Metric name.
- `ui_metric_group` (String) — UI metric group.
- `ui_metric_name` (String) — UI display name for the metric.
- `block_name` (String) — UI block name.

#### `kubernetes_cluster_object_states_results` (List of Object)

- `metric_name` (String) — Metric name.
- `ui_metric_group` (String) — UI metric group.
- `ui_color_group_id` (List of Number) — UI color group IDs.
- `ui_metric_name` (String) — UI display name for the metric.
- `subobject_name` (String) — Subobject name.
- `value` (List of String) — State values.

#### `kubernetes_cluster_current_state_results` (List of Object)

- `subobject_name` (String) — Subobject name.
- `metric_name` (String) — Metric name.
- `ui_metric_group` (String) — UI metric group.
- `ui_color_group_id` (Number) — UI color group ID.
- `ui_metric_name` (String) — UI display name for the metric.
- `value` (Number) — Current state metric value.

#### `kubernetes_cluster_changing_metrics_results` (List of Object)

- `subobject_name` (String) — Subobject name.
- `metric_name` (String) — Metric name.
- `ui_metric_group` (String) — UI metric group.
- `ui_color_group_id` (Number) — UI color group ID.
- `ui_metric_name` (String) — UI display name for the metric.
- `human_label` (String) — Human-readable label.
- `timecodes` (List of String) — List of timecodes.
- `linechart_values` (List of Number) — Values for line chart.
- `treemap_values` (Number) — Value for treemap visualization.
