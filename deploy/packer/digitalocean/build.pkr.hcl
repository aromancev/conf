source "digitalocean" "server" {
  api_token     = var.do_token
  region        = var.region
  image         = "ubuntu-22-04-x64"
  size          = "s-1vcpu-512mb-10gb"
  ssh_username  = "root"
  snapshot_name = var.snapshot_name
}

build {
  sources = ["source.digitalocean.server"]

  provisioner "shell" {
    script = "./setup.sh"
  }
}
