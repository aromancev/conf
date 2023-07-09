terraform {
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
