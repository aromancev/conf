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

# Use `openssl rand -base64 32` to generate.
variable "consul_gossip_key" {
  description = "Gossip key for Consul."
  type        = string
  sensitive   = true
}

# Use `openssl rand -base64 32` to generate.
variable "nomad_gossip_key" {
  description = "Gossip key for Nomad."
  type        = string
  sensitive   = true
}

# This could be generated in Terraform using https://registry.terraform.io/providers/hashicorp/tls
# but it doesn't set the Autority on the certificate correctly: https://github.com/hashicorp/terraform-provider-tls/pull/309
# Use `secrets.sh` to generate.
variable "consul_ca_cert_pem" {
  description = "Consul CA (consul-agent-ca.pem)."
  type        = string
  sensitive   = true
}

# This could be generated in Terraform using https://registry.terraform.io/providers/hashicorp/tls
# but it doesn't set the Autority on the certificate correctly: https://github.com/hashicorp/terraform-provider-tls/pull/309
variable "consul_server_cert_pem" {
  description = "Consul server certificate (dc1-server-consul-0.pem)."
  type        = string
  sensitive   = true
}

# This could be generated in Terraform using https://registry.terraform.io/providers/hashicorp/tls
# but it doesn't set the Autority on the certificate correctly: https://github.com/hashicorp/terraform-provider-tls/pull/309
variable "consul_server_private_key_pem" {
  description = "Consul server private key (dc1-server-consul-0-key.pem)"
  type        = string
  sensitive   = true
}

# This could be generated in Terraform using https://registry.terraform.io/providers/hashicorp/tls
# but it doesn't set the Autority on the certificate correctly: https://github.com/hashicorp/terraform-provider-tls/pull/309
variable "nomad_ca_cert_pem" {
  description = "Nomad CA (nomad-agent-ca.pem)."
  type        = string
  sensitive   = true
}

# This could be generated in Terraform using https://registry.terraform.io/providers/hashicorp/tls
# but it doesn't set the Autority on the certificate correctly: https://github.com/hashicorp/terraform-provider-tls/pull/309
variable "nomad_server_cert_pem" {
  description = "Nomad server certificate (global-server-nomad.pem)."
  type        = string
  sensitive   = true
}

# This could be generated in Terraform using https://registry.terraform.io/providers/hashicorp/tls
# but it doesn't set the Autority on the certificate correctly: https://github.com/hashicorp/terraform-provider-tls/pull/309
variable "nomad_server_private_key_pem" {
  description = "Nomad server key (global-server-nomad-key.pem)"
  type        = string
  sensitive   = true
}

# This could be generated in Terraform using https://registry.terraform.io/providers/hashicorp/tls
# but it doesn't set the Autority on the certificate correctly: https://github.com/hashicorp/terraform-provider-tls/pull/309
variable "nomad_client_cert_pem" {
  description = "Nomad client certificate (global-client-nomad.pem)."
  type        = string
  sensitive   = true
}

# This could be generated in Terraform using https://registry.terraform.io/providers/hashicorp/tls
# but it doesn't set the Autority on the certificate correctly: https://github.com/hashicorp/terraform-provider-tls/pull/309
variable "nomad_client_private_key_pem" {
  description = "Nomad client private key (global-client-nomad-key.pem)"
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
