resource "emma_security_group" "security_group" {
  name  = "Example"
  rules = [
    {
      direction = "INBOUND"
      protocol  = "all"
      ports     = "8080"
      ip_range  = "8.8.8.8/32"
    },
    {
      direction = "INBOUND"
      protocol  = "all"
      ports     = "8080"
      ip_range  = "4.4.4.4/32"
    }
  ]
}