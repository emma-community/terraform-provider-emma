resource "emma_vm" "vm" {
  name               = "vm-test1"
  data_center_id     = "data_center_id"
  os_id              = "os_id"
  cloud_network_type = "multi-cloud"
  vcpu_type          = "shared"
  vcpu               = 2
  ram_gb             = 1
  volume_type        = "ssd"
  volume_gb          = 8
  ssh_key_id         = 570
}