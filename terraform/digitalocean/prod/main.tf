provider "digitalocean" {
  token = var.digitalocean_token
}

resource "digitalocean_project" "main" {
  name        = "Confa"
  description = "Production confa.io project."
  purpose     = "Web Application"
  environment = "Production"

  resources = [
    digitalocean_droplet.main.urn,
  ]
}

resource "digitalocean_vpc" "main" {
  name     = "confa"
  region   = "ams3"
  ip_range = "10.10.10.0/24"
}

resource "digitalocean_ssh_key" "main" {
  name       = "Confa Terraform"
  public_key = file(var.ssh_key_public)
}

resource "digitalocean_droplet" "main" {
  image    = "ubuntu-22-04-x64"
  name     = "main"
  region   = "ams3"
  size     = "s-1vcpu-1gb"
  vpc_uuid = digitalocean_vpc.main.id
  ssh_keys = [
    digitalocean_ssh_key.main.fingerprint,
  ]

  connection {
    host        = self.ipv4_address
    user        = "root"
    type        = "ssh"
    private_key = file(var.ssh_key_private)
    timeout     = "2m"
  }

  provisioner "remote-exec" {
    inline = [
      "ufw default deny incoming",
      "ufw default allow outgoing",
      "ufw allow 80/tcp",
      "ufw allow 443/tcp",
      "ufw --force enable",
    ]
  }
}

resource "digitalocean_droplet" "sfu" {
  image    = "ubuntu-22-04-x64"
  name     = "sfu"
  region   = "ams3"
  size     = "s-1vcpu-1gb"
  vpc_uuid = digitalocean_vpc.main.id
  ssh_keys = [
    digitalocean_ssh_key.main.fingerprint,
  ]

  connection {
    host        = self.ipv4_address
    user        = "root"
    type        = "ssh"
    private_key = file(var.ssh_key_private)
    timeout     = "2m"
  }

  provisioner "remote-exec" {
    inline = [
      "ufw default deny incoming",
      "ufw default allow outgoing",
      "ufw allow 80/tcp",
      "ufw allow 443/tcp",
      "ufw allow 50000:60000/udp",
      "ufw --force enable",
    ]
  }
}
