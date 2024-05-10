resource "emma_spot_instance" "spot_instance" {
  name               = "Example"
  data_center_id     = data.emma_data_center.aws_spot.id
  os_id              = data.emma_operating_system.ubuntu.id
  cloud_network_type = "multi-cloud"
  vcpu_type          = "shared"
  vcpu               = 2
  ram_gb             = 1
  volume_type        = "ssd"
  volume_gb          = 8
  security_group_id  = emma_security_group.security_group.id
  ssh_key_id         = emma_ssh_key.ssh_key.id
  price              = 0.00305205479452
}