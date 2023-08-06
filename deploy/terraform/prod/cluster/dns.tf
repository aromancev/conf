resource "cloudflare_record" "root" {
  zone_id = var.cloudflare_zone_id
  name    = "@"
  value   = digitalocean_reserved_ip.ingress.ip_address
  type    = "A"
  ttl     = 3600
}

resource "cloudflare_record" "sfu" {
  zone_id = var.cloudflare_zone_id
  name    = "sfu"
  value   = digitalocean_reserved_ip.ingress.ip_address
  type    = "A"
  ttl     = 3600
}

resource "cloudflare_record" "turn" {
  zone_id = var.cloudflare_zone_id
  name    = "turn"
  value   = digitalocean_reserved_ip.ingress.ip_address
  type    = "A"
  ttl     = 3600
}

resource "cloudflare_record" "teleport" {
  zone_id = var.cloudflare_zone_id
  name    = "*.teleport"
  value   = digitalocean_reserved_ip.ops.ip_address
  type    = "A"
  ttl     = 3600
}
