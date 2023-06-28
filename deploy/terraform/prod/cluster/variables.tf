variable "do_token" {
  description = "DigitalOcean API token."
  type        = string
  sensitive   = true
}

variable "do_token_cloud_autoconnect" {
  description = "DigitalOcean API token for cloud autoconnect. More info: https://github.com/hashicorp/go-discover."
  type        = string
  sensitive   = true
}

variable "region" {
  description = "DigitalOcean region to deploy the cluter."
  type        = string
}

variable "ssh_key_public" {
  description = "File path to public SSH key that will be used to access droplets."
  type        = string
}

variable "ssh_key_private" {
  description = "File path to private SSH key that will be used to access droplets."
  type        = string
}

variable "github_actions_repo" {
  description = "GitHub respository URL for self-hosted runner."
  type        = string
}

variable "github_actions_token" {
  description = "GitHub runner registration token."
  type        = string
  sensitive   = true
}

variable "datacenter" {
  description = "Cluster datacenter name."
  type        = string
  default     = "dc1"
}
