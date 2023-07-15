resource "tls_private_key" "main" {
  algorithm   = "ECDSA"
  ecdsa_curve = "P256"
}
