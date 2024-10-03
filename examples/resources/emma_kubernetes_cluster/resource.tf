resource "emma_kubernetes_cluster" "kubernetes_cluster" {
  name                = "example"
  deployment_location = "eu"
  domain_name         = null
  worker_nodes = [
    {
      name           = "example-worker-node"
      data_center_id = "gcp-europe-west8-a"
      vcpu_type      = "shared"
      vcpu           = 2
      ram_gb         = 2
      volume_type    = "ssd"
      volume_gb      = 16
    },
  ]
  autoscaling_configs = [
    {
      group_name     = "example-autoscaling-group"
      data_center_id = "gcp-europe-west8-a"
      minimum_nodes  = 1
      maximum_nodes  = 2
      target_nodes   = 1

      node_group_price_limit                   = 5
      use_on_demand_instances_instead_of_spots = false
      spot_percent                             = 40
      spot_markup                              = 3
      configuration_priorities = [
        {
          vcpu_type   = "shared"
          volume_type = "ssd"
          vcpu        = 2
          ram_gb      = 4
          volume_gb   = 16
          priority    = "high"
        }
      ]
    },
  ]
}
