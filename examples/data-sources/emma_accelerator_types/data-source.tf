# List every GPU accelerator type exposed by the platform.
data "emma_accelerator_types" "all" {}

output "available_gpus" {
  value = [for t in data.emma_accelerator_types.all.accelerator_types : t.accelerator_type]
}

# Pick one and use it to filter VM configurations.
data "emma_vm_configurations" "gpu_configs" {
  accelerator_type_id = data.emma_accelerator_types.all.accelerator_types[0].id
}
