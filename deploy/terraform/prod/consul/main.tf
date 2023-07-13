resource "tls_private_key" "main" {
  algorithm   = "ECDSA"
  ecdsa_curve = "P256"
}

resource "consul_key_prefix" "auth" {
  path_prefix = "auth/"

  subkeys = {
    "private_key" = tls_private_key.main.private_key_pem
    "public_key"  = tls_private_key.main.public_key_pem
  }
}
