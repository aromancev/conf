job "traefik" {
  type = "system"
  datacenters = ["dc1"]

  group "traefik" {
    network {
      port "http" {
        static = 80
      }
      port "api" {
        static = 8080
      }
    }

    service {
      name = "traefik"

      check {
        name     = "alive"
        type     = "tcp"
        port     = "http"
        interval = "10s"
        timeout  = "2s"
      }
    }

    task "traefik" {
      driver = "docker"

      config {
        image        = "traefik:v2.2"
        network_mode = "host"

        volumes = [
          "local/traefik.yml:/etc/traefik/traefik.yml",
        ]
      }

      template {
        data = <<EOF
entryPoints:
  http:
    address: ":80"
  traefik:
    address: ":8080"

api:
  dashboard: true
  insecure: true

providers:
  consulCatalog:
    prefix: "traefik"
    exposedByDefault: false
    endpoint:
      address: "127.0.0.1:8500"
      scheme: "http"
EOF
        destination = "local/traefik.yml"
      }

      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}
