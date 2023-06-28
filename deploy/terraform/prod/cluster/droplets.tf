locals {
  bootstrap_expect = 1
  vpc_range        = "10.10.10.0/24"
  mongo_group_id   = "1001"
  mongo_user_id    = "1001"
}

resource "digitalocean_ssh_key" "main" {
  name       = "Confa Terraform"
  public_key = file(var.ssh_key_public)
}

data "digitalocean_droplet_snapshot" "cluster_server" {
  name_regex  = "^confa_cluster_server$"
  region      = var.region
  most_recent = true
}

data "digitalocean_droplet_snapshot" "cluster_worker" {
  name_regex  = "^confa_cluster_worker$"
  region      = var.region
  most_recent = true
}

resource "digitalocean_vpc" "main" {
  name     = "confa-vpc"
  region   = var.region
  ip_range = local.vpc_range
}

resource "digitalocean_droplet" "server" {
  image    = data.digitalocean_droplet_snapshot.cluster_server.id
  name     = "cluster-server"
  region   = var.region
  size     = "s-1vcpu-512mb-10gb"
  tags     = ["consul-autojoin"]
  vpc_uuid = digitalocean_vpc.main.id
  ssh_keys = [
    digitalocean_ssh_key.main.fingerprint,
  ]

  # If `user_data` doesn't work, check cloud init logs on the instance: /var/log/cloud-init-output.log
  user_data = <<EOT
#!/bin/bash -e

# Setup Consul.
echo 'Creating Consul config files in  /etc/consul.d/ ...'
mkdir --parents /etc/consul.d

echo '
datacenter = "${var.datacenter}"
retry_join = ["provider=digitalocean region=${var.region} tag_name=consul-autojoin api_token=${var.do_token_cloud_autoconnect}"]

# Address to communication inside Consul cluster. Only bind to the private IP on the VPC.
bind_addr = "{{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"address\" }}"

data_dir = "/opt/consul"

server = true
bootstrap_expect = ${local.bootstrap_expect}

addresses {
  # Address for Consul client. Bind to private network so that other agents can communicate inside the cluster.
  http = "127.0.0.1 {{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"address\" }}"
  dns = "127.0.0.1"
  grpc = "127.0.0.1"
}

ports {
  # Allow unsecured HTTP traffic.
  # This is safe if Consul is exposed ONLY on localhost. Otherwise set to -1.
  http = 8500
}
' > /etc/consul.d/consul.hcl

echo 'Starting Consul service ...'
chown --recursive consul:consul /etc/consul.d
chmod 700 /etc/consul.d
systemctl enable consul
systemctl start consul

# Setup Nomad.
echo 'Creating Nomad config files in  /etc/nomad.d/ ...'
mkdir --parents /etc/nomad.d

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

server {
  enabled = true
  bootstrap_expect = ${local.bootstrap_expect}
}
' > /etc/nomad.d/nomad.hcl

echo 'Starting Nomad service ...'
chown --recursive nomad:nomad /etc/nomad.d
chmod 700 /etc/nomad.d
systemctl enable nomad
systemctl start nomad

echo 'Bootstrap complete!'
  EOT
}

resource "digitalocean_droplet" "ingress" {
  image    = data.digitalocean_droplet_snapshot.cluster_worker.id
  name     = "cluster-ingress"
  region   = var.region
  size     = "s-1vcpu-1gb"
  tags     = ["consul-autojoin"]
  vpc_uuid = digitalocean_vpc.main.id
  ssh_keys = [
    digitalocean_ssh_key.main.fingerprint,
  ]

  # If `user_data` doesn't work, check cloud init logs on the instance: /var/log/cloud-init-output.log
  user_data = <<EOT
#!/bin/bash -e

# Setup Consul.
echo 'Creating Consul config files in  /etc/consul.d/ ...'
mkdir --parents /etc/consul.d

echo '
datacenter = "${var.datacenter}"
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
}

ui_config {
  enabled = true
}
' > /etc/consul.d/consul.hcl

echo 'Starting Consul service ...'
chown --recursive consul:consul /etc/consul.d
chmod 700 /etc/consul.d
systemctl enable consul
systemctl start consul

# Setup Nomad.
echo 'Creating Nomad config files in  /etc/nomad.d/ ...'
mkdir --parents /etc/nomad.d

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

client {
  enabled = true
  network_interface = "{{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"name\" }}"
}

ui {
  enabled = true
}

# This allows Nomad to register as a service with the Consul agent on the same host.
consul {
  address = "127.0.0.1:8500"
}
' > /etc/nomad.d/nomad.hcl

echo 'Starting Nomad service ...'
chown --recursive nomad:nomad /etc/nomad.d
chmod 700 /etc/nomad.d
systemctl enable nomad
systemctl start nomad

echo 'Configuring Consul DNS forwarding ...'
# Clear all systemd-resolved configs. DigitalOcean puts it's own servers there. 
# We want to replace it with our Consul instance running locally. More info: https://developer.hashicorp.com/consul/tutorials/networking/dns-forwarding.
rm /etc/systemd/resolved.conf.d/*
# Create Consul DNS forwarding.
echo '
[Resolve]
DNS=127.0.0.1:8600
DNSSEC=false
Domains=~consul
' > /etc/systemd/resolved.conf.d/consul.conf
systemctl restart systemd-resolved

echo 'Bootstrap complete!'
  EOT
}

resource "digitalocean_droplet" "ops" {
  image    = data.digitalocean_droplet_snapshot.cluster_worker.id
  name     = "cluster-ops"
  region   = var.region
  size     = "s-1vcpu-1gb"
  tags     = ["consul-autojoin"]
  vpc_uuid = digitalocean_vpc.main.id
  ssh_keys = [
    digitalocean_ssh_key.main.fingerprint,
  ]

  # If `user_data` doesn't work, check cloud init logs on the instance: /var/log/cloud-init-output.log
  user_data = <<EOT
#!/bin/bash -e

# Setup Consul.
echo 'Creating Consul config files in  /etc/consul.d/ ...'
mkdir --parents /etc/consul.d

echo '
datacenter = "${var.datacenter}"
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
}

ui_config {
  enabled = true
}
' > /etc/consul.d/consul.hcl

echo 'Starting Consul service ...'
chown --recursive consul:consul /etc/consul.d
chmod 700 /etc/consul.d
systemctl enable consul
systemctl start consul

# Setup Nomad.
echo 'Creating Nomad config files in  /etc/nomad.d/ ...'
mkdir --parents /etc/nomad.d

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

client {
  enabled = true
  network_interface = "{{ GetPrivateInterfaces | include \"network\" \"${local.vpc_range}\" | attr \"name\" }}"

  # Host mount that the mongo job can use. It is also used by Nomad to select the node for scheduling.
  host_volume "mongodb" {
    path      = "/opt/mongodb"
    read_only = false
  }
}

ui {
  enabled = true
}

# This allows Nomad to register as a service with the Consul agent on the same host.
consul {
  address = "127.0.0.1:8500"
}
' > /etc/nomad.d/nomad.hcl

# Creating host directory for MongoDB.
mkdir --parents /opt/mongodb/data
chmod 700 /opt/mongodb/data
# Hack for enabling replication on a single instance.
# TODO: handle roperly in case of multiple instances.
openssl rand -base64 32 > /opt/mongodb/repl.key
chmod 400 /opt/mongodb/repl.key
groupadd -g ${local.mongo_group_id} mongodb
useradd -M -s /bin/false -g mongodb -u ${local.mongo_user_id} mongodb
chown --recursive mongodb:mongodb /opt/mongodb

echo 'Starting Nomad service ...'
chown --recursive nomad:nomad /etc/nomad.d
chmod 700 /etc/nomad.d
systemctl enable nomad
systemctl start nomad

echo 'Configuring GitHub Actions runner ...'
cd /home/github
sudo -H -u github bash -c './config.sh --unattended --replace --url ${var.github_actions_repo} --token ${var.github_actions_token}'
./svc.sh install github
./svc.sh start

echo 'Configuring Consul DNS forwarding ...'
# Clear all systemd-resolved configs. DigitalOcean puts it's own servers there. 
# We want to replace it with our Consul instance running locally. More info: https://developer.hashicorp.com/consul/tutorials/networking/dns-forwarding.
rm /etc/systemd/resolved.conf.d/*
# Create Consul DNS forwarding.
echo '
[Resolve]
DNS=127.0.0.1:8600
DNSSEC=false
Domains=~consul
' > /etc/systemd/resolved.conf.d/consul.conf
systemctl restart systemd-resolved

echo 'Bootstrap complete!'
  EOT
}
