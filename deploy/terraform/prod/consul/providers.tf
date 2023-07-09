provider "consul" {
  address    = var.consul_host
  datacenter = var.datacenter
}
