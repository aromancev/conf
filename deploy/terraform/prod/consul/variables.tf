variable "consul_host" {
  description = "Consul host address."
  type        = string
}

variable "datacenter" {
  description = "Cluster datacenter name."
  type        = string
}

variable "domain" {
  description = "Domain to use for the main website."
  type        = string
}

variable "cert_email" {
  description = "Email to use in the TLS certificate administration."
  type        = string
}

variable "mailersend_token" {
  description = "MailerSend API token."
  type        = string
  sensitive   = true
}

variable "google_client_id" {
  description = "Client ID for Google authentication."
  type        = string
}

variable "google_client_secret" {
  description = "Client Secret for Google authentication."
  type        = string
  sensitive   = true
}

variable "storage_access_key" {
  description = "Access key for S3 compatibe object storage."
  type        = string
}

variable "storage_secret_key" {
  description = "Secret key for S3 compatibe object storage."
  type        = string
  sensitive   = true
}
