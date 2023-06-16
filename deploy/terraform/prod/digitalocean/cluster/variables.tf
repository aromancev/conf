variable "do_token" {
  description = "DigitalOcean API token."
  type        = string
  sensitive   = true
}

variable "region" {
  description = "DigitalOcean region to deploy the cluter."
  type        = string
}

variable "do_token_cloud_autoconnect" {
  description = "DigitalOcean API token for cloud autoconnect. More info: https://github.com/hashicorp/go-discover."
  type        = string
  sensitive   = true
}

variable "ssh_key_public" {
  description = "File path to public SSH key that will be used to access droplets."
  type        = string
}

variable "ssh_key_private" {
  description = "File path to private SSH key that will be used to access droplets."
  type        = string
}

variable "consul_gossip_key" {
  description = "Gossip key for Consul (`openssl rand -base64 32`)."
  type        = string
  sensitive   = true
}

variable "nomad_gossip_key" {
  description = "Gossip key for Nomad. (`openssl rand -base64 32`)."
  type        = string
  sensitive   = true
}

variable "consul_ca_cert" {
  description = ""
  type        = string
  sensitive   = true
}

variable "consul_server_cert" {
  description = ""
  type        = string
  sensitive   = true
}

variable "consul_server_private_key" {
  description = ""
  type        = string
  sensitive   = true
}

variable "server_count" {
  description = "Number of server instances. Those host cluster servers such as Consul, Nomad, Vault, etc."
  type        = number
  default     = 1
}

variable "worker_count" {
  description = "Number of worker instances. Those will be used to allocate Nomad jobs (essentially, where applications will live)."
  type        = number
  default     = 1
}

variable "datacenter" {
  description = "Cluster datacenter name."
  type        = string
  default     = "dc1"
}
