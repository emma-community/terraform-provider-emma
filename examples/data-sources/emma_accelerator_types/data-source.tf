# List every GPU accelerator type exposed by the platform.
data "emma_accelerator_types" "all" {}

output "nvidia_t4_id" {
  value = data.emma_accelerator_types.all.ids_by_name["NVIDIA T4"]
}

output "available_gpu_names" {
  value = keys(data.emma_accelerator_types.all.ids_by_name)
}

output "available_gpus" {
  value = data.emma_accelerator_types.all.accelerator_types
}

output "available_gpus_map" {
  value = data.emma_accelerator_types.all.ids_by_name
}

# Feed an id into another data source to filter VM configurations by GPU.
data "emma_vm_configurations" "gpu_t4_configs" {
  accelerator_type_id = data.emma_accelerator_types.all.ids_by_name["NVIDIA T4"]
}
