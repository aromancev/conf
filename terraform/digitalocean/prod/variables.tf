variable "digitalocean_token" {
  description = "DigitalOcean API token."
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
