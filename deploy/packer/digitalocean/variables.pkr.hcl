variable "do_token" {
  description = "DigitalOcean API token."
  type        = string
  sensitive   = true
}

variable "region" {
  description = "DigitalOcean region to place the snapshot in."
  type        = string
}

variable "snapshot_name" {
  description = "Name of the snapshot. This is used to find the snapshot programmaticaly (using Terraform, for example)."
  type        = string
}
