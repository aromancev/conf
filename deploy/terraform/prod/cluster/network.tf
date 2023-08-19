locals {
  vpc_range = "10.10.10.0/24"
}

resource "digitalocean_reserved_ip" "ingress" {
  region = var.region
}

resource "digitalocean_reserved_ip" "ops" {
  region = var.region
}

resource "digitalocean_reserved_ip_assignment" "ingress" {
  ip_address = digitalocean_reserved_ip.ingress.ip_address
  droplet_id = digitalocean_droplet.ingress.id
}

resource "digitalocean_reserved_ip_assignment" "ops" {
  ip_address = digitalocean_reserved_ip.ops.ip_address
  droplet_id = digitalocean_droplet.ops.id
}

resource "digitalocean_vpc" "main" {
  name     = "confa-vpc"
  region   = var.region
  ip_range = local.vpc_range
}
