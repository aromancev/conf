job "web" {
  datacenters = ["dc1"]

  group "web" {
    network {
      port "web" {
        to = 80
      }
    }

    service {
      name = "web"
      port = "web"

      tags = [
        "traefik.enable=true",
        "traefik.http.routers.web.rule=PathPrefix(`/`)",
      ]

      check {
        name     = "alive"
        type     = "http"
        path     = "/"
        interval = "2s"
        timeout  = "2s"
      }
    }

    task "web" {
      driver = "docker"

      config {
        image = "confa/web:latest"
        ports = ["web"]
      }

      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}
