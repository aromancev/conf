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

echo '${var.consul_ca_cert_pem}' > /etc/consul.d/certs/consul-agent-ca.pem
echo '${var.consul_server_cert_pem}' > /etc/consul.d/certs/dc1-server-consul-0.pem
echo '${var.consul_server_private_key_pem}' > /etc/consul.d/certs/dc1-server-consul-0-key.pem

echo '
datacenter = "${var.datacenter}"
encrypt = "${var.consul_gossip_key}"
retry_join = ["provider=digitalocean region=${var.region} tag_name=consul-autojoin api_token=${var.do_token_cloud_autoconnect}"]

# Address to communication inside Consul cluster. Only bind to the private IP on the VPC.
bind_addr = "{{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"address\" }}"

data_dir = "/opt/consul"

server = true
bootstrap_expect = ${local.bootstrap_expect}

addresses {
  # Address for Consul client. Bind to localhost so that other services on the same host can register themselves.
  http = "127.0.0.1"
  # Address for Consul client. Bind to localhost so that other agents can communicate inside the cluster.
  https = "{{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"address\" }}"
  dns = "127.0.0.1"
  grpc = "127.0.0.1"
  grpc_tls = "127.0.0.1"
}

ports {
  # Allow unsecured HTTP traffic.
  # This is safe if Consul is exposed ONLY on localhost. Otherwise set to -1.
  http = 8500
  https = 8501
  grpc_tls  = 8503
}

tls {
  defaults {
    ca_file = "/etc/consul.d/certs/consul-agent-ca.pem"
    cert_file = "/etc/consul.d/certs/${var.datacenter}-server-consul-0.pem"
    key_file = "/etc/consul.d/certs/${var.datacenter}-server-consul-0-key.pem"

    verify_incoming = true
    verify_outgoing = true
  }

  internal_rpc {
    verify_server_hostname = true
  }
}

auto_encrypt {
  allow_tls = true
}

connect {
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
sudo mkdir --parents /etc/nomad.d/certs

echo '${var.nomad_ca_cert_pem}' > /etc/consul.d/certs/nomad-agent-ca.pem
echo '${var.nomad_server_cert_pem}' > /etc/consul.d/certs/global-server-nomad.pem
echo '${var.nomad_server_private_key_pem}' > /etc/consul.d/certs/global-server-nomad-key.pem

echo '
datacenter = "${var.datacenter}"
data_dir = "/opt/nomad"

addresses {
  # Address for Nomad client. Only bind to localhost to allow access from services running on the same machine.
  http = "127.0.0.1"
  rpc = "{{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"address\" }}"
  serf = "{{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"address\" }}"
}

advertise {
  http = "127.0.0.1"
  rpc = "{{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"address\" }}"
  serf = "{{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"address\" }}"
}

# This allows Nomad to register as a service with the Consul agent on the same host.
consul {
  address = "127.0.0.1:8500"
}

tls {
  # Allow unsecured HTTP traffic.
  # This is safe if Nomad is exposed ONLY on localhost. Otherwise set to false.
  http = false
  rpc  = true

  ca_file   = "/etc/consul.d/certs/nomad-agent-ca.pem"
  cert_file = "/etc/consul.d/certs/global-server-nomad.pem"
  key_file  = "/etc/consul.d/certs/global-server-nomad-key.pem"

  verify_server_hostname = true
  verify_https_client    = true
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

echo '${var.consul_ca_cert_pem}' > /etc/consul.d/certs/consul-agent-ca.pem

echo '
datacenter = "${var.datacenter}"
encrypt = "${var.consul_gossip_key}"
retry_join = ["provider=digitalocean region=${var.region} tag_name=consul-autojoin api_token=${var.do_token_cloud_autoconnect}"]

# Address to communication inside Consul cluster. Only bind to the private IP.
bind_addr = "{{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"address\" }}"
# Address for Consul client. Bind localhost so that services like Nomad on the same host can register themselves. 
# Also bind to the public IP to serve the UI. 
client_addr = "127.0.0.1 {{ GetPublicInterfaces | attr \"address\" }}"

data_dir = "/opt/consul"

ports {
  # Allow unsecured HTTP traffic.
  # This is safe if Consul is exposed ONLY on localhost. Otherwise set to -1.
  http = 8500
  https = 8501
}

tls {
  defaults {
    ca_file = "/etc/consul.d/certs/consul-agent-ca.pem"

    verify_incoming = true
    verify_outgoing = true
  }

  internal_rpc {
    verify_server_hostname = true
  }
}

auto_encrypt = {
  tls = true
}

ui_config {
  enabled = true
}
' > /etc/consul.d/consul.hcl

echo 'Starting Consul service ...'
sudo chown --recursive consul:consul /etc/consul.d
sudo chmod 700 /etc/consul.d
sudo systemctl enable consul
sudo systemctl start consul

# Setup Nomad.
echo 'Creating Nomad config files in  /etc/nomad.d/ ...'
sudo mkdir --parents /etc/nomad.d/certs

echo '${var.nomad_ca_cert_pem}' > /etc/consul.d/certs/nomad-agent-ca.pem
echo '${var.nomad_client_cert_pem}' > /etc/consul.d/certs/global-client-nomad.pem
echo '${var.nomad_client_private_key_pem}' > /etc/consul.d/certs/global-client-nomad-key.pem

echo '
datacenter = "${var.datacenter}"
data_dir = "/opt/nomad"

addresses {
  # Address for Nomad client. Only bind to localhost to allow access from services running on the same machine.
  http = "127.0.0.1 {{ GetPublicInterfaces | attr \"address\" }}"
  rpc = "{{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"address\" }}"
  serf = "{{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"address\" }}"
}

advertise {
  http = "{{ GetPublicInterfaces | attr \"address\" }}"
  rpc = "{{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"address\" }}"
  serf = "{{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"address\" }}"
}

ui {
  enabled = true
}

# This allows Nomad to register as a service with the Consul agent on the same host.
consul {
  address = "127.0.0.1:8500"
}

tls {
  # Allow unsecured HTTP traffic.
  # This is safe if Nomad is exposed ONLY on localhost. Otherwise set to true.
  http = false
  rpc  = true

  ca_file   = "/etc/consul.d/certs/nomad-agent-ca.pem"
  cert_file = "/etc/consul.d/certs/global-client-nomad.pem"
  key_file  = "/etc/consul.d/certs/global-client-nomad-key.pem"

  verify_server_hostname = true
  verify_https_client    = true
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
