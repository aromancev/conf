terraform {
  backend "consul" {}

  required_providers {
    consul = {
      source  = "hashicorp/consul"
      version = "~> 2.17"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }

  required_version = "~> 1.4"
}

provider "consul" {
  address    = var.consul_host
  datacenter = var.datacenter
}
