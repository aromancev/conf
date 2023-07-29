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
