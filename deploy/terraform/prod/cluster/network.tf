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

# resource "digitalocean_firewall" "ingress" {
#   name = "ingress"

#   droplet_ids = [digitalocean_droplet.ingress.id]

#   # Allow SSH.
#   inbound_rule {
#     protocol         = "tcp"
#     port_range       = "22"
#     source_addresses = ["0.0.0.0/0", "::/0"]
#   }
#   # Allow HTTP.
#   inbound_rule {
#     protocol         = "tcp"
#     port_range       = "80"
#     source_addresses = ["0.0.0.0/0", "::/0"]
#   }
#   # Allow HTTPS.
#   inbound_rule {
#     protocol         = "tcp"
#     port_range       = "443"
#     source_addresses = ["0.0.0.0/0", "::/0"]
#   }

#   # Allow SFU.
#   inbound_rule {
#     protocol         = "udp"
#     port_range       = "50000-60000"
#     source_addresses = ["0.0.0.0/0", "::/0"]
#   }

#   # Allow SFU RTC.
#   inbound_rule {
#     protocol         = "tcp"
#     port_range       = "7881"
#     source_addresses = ["0.0.0.0/0", "::/0"]
#   }

#   # Allow SFU UDP TURN.
#   inbound_rule {
#     protocol         = "udp"
#     port_range       = "3478"
#     source_addresses = ["0.0.0.0/0", "::/0"]
#   }

#   # Allow all traffic from cluster.
#   inbound_rule {
#     protocol           = "tcp"
#     port_range         = "1-65535"
#     source_droplet_ids = local.cluster_droplet_ids
#   }
#   inbound_rule {
#     protocol           = "udp"
#     port_range         = "1-65535"
#     source_droplet_ids = local.cluster_droplet_ids
#   }
#   inbound_rule {
#     protocol           = "icmp"
#     port_range         = "1-65535"
#     source_droplet_ids = local.cluster_droplet_ids
#   }

#   # Allow all outbound traffic.
#   outbound_rule {
#     protocol              = "tcp"
#     port_range            = "1-65535"
#     destination_addresses = ["0.0.0.0/0", "::/0"]
#   }
#   outbound_rule {
#     protocol              = "udp"
#     port_range            = "1-65535"
#     destination_addresses = ["0.0.0.0/0", "::/0"]
#   }
#   outbound_rule {
#     protocol              = "icmp"
#     port_range            = "1-65535"
#     destination_addresses = ["0.0.0.0/0", "::/0"]
#   }
# }

# resource "digitalocean_firewall" "ops" {
#   name = "ops"

#   droplet_ids = [digitalocean_droplet.ops.id]

#   # Allow SSH.
#   inbound_rule {
#     protocol         = "tcp"
#     port_range       = "22"
#     source_addresses = ["0.0.0.0/0", "::/0"]
#   }
#   # Allow HTTP.
#   inbound_rule {
#     protocol         = "tcp"
#     port_range       = "80"
#     source_addresses = ["0.0.0.0/0", "::/0"]
#   }
#   # Allow HTTPS.
#   inbound_rule {
#     protocol         = "tcp"
#     port_range       = "443"
#     source_addresses = ["0.0.0.0/0", "::/0"]
#   }

#   # Allow all traffic from cluster.
#   inbound_rule {
#     protocol           = "tcp"
#     port_range         = "1-65535"
#     source_droplet_ids = local.cluster_droplet_ids
#   }
#   inbound_rule {
#     protocol           = "udp"
#     port_range         = "1-65535"
#     source_droplet_ids = local.cluster_droplet_ids
#   }
#   inbound_rule {
#     protocol           = "icmp"
#     port_range         = "1-65535"
#     source_droplet_ids = local.cluster_droplet_ids
#   }

#   # Allow all outbound traffic.
#   outbound_rule {
#     protocol              = "tcp"
#     port_range            = "1-65535"
#     destination_addresses = ["0.0.0.0/0", "::/0"]
#   }
#   outbound_rule {
#     protocol              = "udp"
#     port_range            = "1-65535"
#     destination_addresses = ["0.0.0.0/0", "::/0"]
#   }
#   outbound_rule {
#     protocol              = "icmp"
#     port_range            = "1-65535"
#     destination_addresses = ["0.0.0.0/0", "::/0"]
#   }
# }

# resource "digitalocean_firewall" "cluster" {
#   name = "intra-cluster"

#   droplet_ids = [digitalocean_droplet.server.id]

#   # Allow all internal traffic.
#   inbound_rule {
#     protocol           = "tcp"
#     port_range         = "1-65535"
#     source_droplet_ids = local.cluster_droplet_ids
#   }
#   inbound_rule {
#     protocol           = "udp"
#     port_range         = "1-65535"
#     source_droplet_ids = local.cluster_droplet_ids
#   }
#   inbound_rule {
#     protocol           = "icmp"
#     port_range         = "1-65535"
#     source_droplet_ids = local.cluster_droplet_ids
#   }

#   # Allow all outbound traffic.
#   outbound_rule {
#     protocol              = "tcp"
#     port_range            = "1-65535"
#     destination_addresses = ["0.0.0.0/0", "::/0"]
#   }
#   outbound_rule {
#     protocol              = "udp"
#     port_range            = "1-65535"
#     destination_addresses = ["0.0.0.0/0", "::/0"]
#   }
#   outbound_rule {
#     protocol              = "icmp"
#     port_range            = "1-65535"
#     destination_addresses = ["0.0.0.0/0", "::/0"]
#   }
# }
