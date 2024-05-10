resource "emma_vm" "vm" {
  name               = "Example"
  data_center_id     = data.emma_data_center.aws.id
  os_id              = data.emma_operating_system.ubuntu.id
  cloud_network_type = "multi-cloud"
  vcpu_type          = "shared"
  vcpu               = 2
  ram_gb             = 1
  volume_type        = "ssd"
  volume_gb          = 8
  ssh_key_id         = 1
}