source "digitalocean" "server" {
  api_token     = var.do_token
  region        = var.region
  image         = "ubuntu-22-04-x64"
  size          = "s-1vcpu-512mb-10gb"
  ssh_username  = "root"
  snapshot_name = "${var.snapshot_name}_server"
}

source "digitalocean" "worker" {
  api_token     = var.do_token
  region        = var.region
  image         = "ubuntu-22-04-x64"
  size          = "s-1vcpu-512mb-10gb"
  ssh_username  = "root"
  snapshot_name = "${var.snapshot_name}_worker"
}

build {
  sources = ["source.digitalocean.server"]

  provisioner "shell" {
    script = "./setup-server.sh"
  }
}

build {
  sources = ["source.digitalocean.worker"]

  provisioner "shell" {
    script = "./setup-worker.sh"
  }
}
