provider "digitalocean" {
  token = var.do_token
}

locals {
  bootstrap_expect = floor(var.server_count / 2) + 1
  vpc_range        = "10.10.10.0/24"
}

data "digitalocean_droplet_snapshot" "cluster" {
  name_regex  = "^confa_cluster$"
  region      = var.region
  most_recent = true
}

resource "digitalocean_ssh_key" "main" {
  name       = "Confa Terraform"
  public_key = file(var.ssh_key_public)
}

resource "digitalocean_vpc" "main" {
  name     = "confa-vpc"
  region   = var.region
  ip_range = local.vpc_range
}

resource "digitalocean_droplet" "server" {
  count    = var.server_count
  image    = data.digitalocean_droplet_snapshot.cluster.id
  name     = "cluster-server-${count.index}"
  region   = var.region
  size     = "s-1vcpu-512mb-10gb"
  tags     = ["consul-autojoin"]
  vpc_uuid = digitalocean_vpc.main.id
  ssh_keys = [
    digitalocean_ssh_key.main.fingerprint,
  ]

  # If `user_data` doesn't work, check cloud init logs on the instance:
  # /var/log/cloud-init.log /var/log/cloud-init-output.log
  user_data = <<EOT
#!/bin/bash -e

# Setup Consul.
echo 'Creating Consul config files in  /etc/consul.d/ ...'
sudo mkdir --parents /etc/consul.d/certs

echo '
datacenter = "${var.datacenter}"
data_dir = "/opt/consul"
encrypt = "${var.consul_gossip_key}"

retry_join = ["provider=digitalocean region=${var.region} tag_name=consul-autojoin api_token=${var.do_token_cloud_autoconnect}"]

server = true
bootstrap_expect = ${local.bootstrap_expect}
# Address to communication inside Consul cluster. Only bind to the private IP on the VPC.
bind_addr = "{{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"address\" }}"
# Address for Consul client. Bind to localhost and private network so that other agents and services can register themselves.
client_addr = "127.0.0.1 {{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"address\" }}"

connect {
  enabled = true
}

ui_config {
  enabled = true
}

performance {
  raft_multiplier = 1
}
' > /etc/consul.d/consul.hcl

echo 'Starting Consul service ...'
sudo chown --recursive consul:consul /etc/consul.d
sudo chmod 700 /etc/consul.d
sudo systemctl enable consul
sudo systemctl start consul

# Setup Nomad.
echo 'Creating Nomad config files in  /etc/nomad.d/ ...'
sudo mkdir --parents /etc/nomad.d

echo '
datacenter = "${var.datacenter}"
data_dir = "/opt/nomad"
# Address for client Nomad communications. Bind to private network so that other agents can register themselves.
bind_addr = "{{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"address\" }}"

# This allows Nomad to register as a service with the Consul agent on the same host.
consul {
  address = "127.0.0.1:8500"
}

server {
  enabled = true
  bootstrap_expect = ${local.bootstrap_expect}
  encrypt = "${var.nomad_gossip_key}"
}
' > /etc/nomad.d/nomad.hcl

echo 'Starting Nomad service ...'
sudo chown --recursive nomad:nomad /etc/nomad.d
sudo chmod 700 /etc/nomad.d
sudo systemctl enable nomad
sudo systemctl start nomad

echo 'Bootstrap complete!'
  EOT
}

resource "digitalocean_droplet" "worker" {
  count    = var.server_count
  image    = data.digitalocean_droplet_snapshot.cluster.id
  name     = "cluster-worker-${count.index}"
  region   = var.region
  size     = "s-1vcpu-512mb-10gb"
  tags     = ["consul-autojoin"]
  vpc_uuid = digitalocean_vpc.main.id
  ssh_keys = [
    digitalocean_ssh_key.main.fingerprint,
  ]

  # If `user_data` doesn't work, check cloud init logs on the instance:
  # /var/log/cloud-init.log /var/log/cloud-init-output.log
  user_data = <<EOT
#!/bin/bash -e

# Setup Consul.
echo 'Creating Consul config files in  /etc/consul.d/ ...'
sudo mkdir --parents /etc/consul.d/certs

echo '
datacenter = "${var.datacenter}"
data_dir = "/opt/consul"
encrypt = "${var.consul_gossip_key}"

# Address to communication inside Consul cluster. Only bind to the private IP.
bind_addr = "{{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"address\" }}"
# Address for Consul client. Bind localhost so that services like Nomad on the same host can register themselves. 
# Also bind to the public IP to serve the UI. 
client_addr = "127.0.0.1 {{ GetPublicInterfaces | attr \"address\" }}"

ui_config {
  enabled = true
}

retry_join = ["provider=digitalocean region=${var.region} tag_name=consul-autojoin api_token=${var.do_token_cloud_autoconnect}"]
' > /etc/consul.d/consul.hcl

echo 'Starting Consul service ...'
sudo chown --recursive consul:consul /etc/consul.d
sudo chmod 700 /etc/consul.d
sudo systemctl enable consul
sudo systemctl start consul

# Setup Nomad.
echo 'Creating Nomad config files in  /etc/nomad.d/ ...'
sudo mkdir --parents /etc/nomad.d

echo '
datacenter = "${var.datacenter}"
data_dir = "/opt/nomad"
# Address for client Nomad communications. Only bind to the public IP to serve the UI.
bind_addr = "{{ GetPublicInterfaces | attr \"address\" }}"

ui {
  enabled = true
}

# This allows Nomad to register as a service with the Consul agent on the same host.
consul {
  address = "127.0.0.1:8500"
}

client {
  enabled = true
}
' > /etc/nomad.d/nomad.hcl

echo 'Starting Nomad service ...'
sudo chown --recursive nomad:nomad /etc/nomad.d
sudo chmod 700 /etc/nomad.d
sudo systemctl enable nomad
sudo systemctl start nomad

echo 'Bootstrap complete!'
  EOT
}
