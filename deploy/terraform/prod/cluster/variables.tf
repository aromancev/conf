variable "do_token" {
  description = "DigitalOcean API token."
  type        = string
  sensitive   = true
}

variable "cloudflare_token" {
  description = "Cloudflare API token."
  type        = string
  sensitive   = true
}

variable "do_token_cloud_autoconnect" {
  description = "DigitalOcean API token for cloud autoconnect. More info: https://github.com/hashicorp/go-discover."
  type        = string
  sensitive   = true
}

variable "cloudflare_zone_id" {
  description = "Cloudflare zone ID to manage DNS for."
  type        = string
}

variable "region" {
  description = "DigitalOcean region to deploy the cluter."
  type        = string
}

variable "datacenter" {
  description = "Cluster datacenter name."
  type        = string
}

variable "github_actions_repo" {
  description = "GitHub respository URL for self-hosted runner."
  type        = string
}

variable "github_actions_pat" {
  description = "GitHub runner PAT to create registration token."
  type        = string
  sensitive   = true
}
