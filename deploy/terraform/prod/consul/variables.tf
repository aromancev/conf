variable "consul_host" {
  description = "Consul host address."
  type        = string
}

variable "datacenter" {
  description = "Cluster datacenter name."
  type        = string
}

variable "mailersend_token" {
  description = "MailerSend API token."
  type        = string
  sensitive   = true
}
