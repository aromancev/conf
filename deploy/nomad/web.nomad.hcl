job "web" {
  datacenters = ["dc1"]

  group "web" {
    network {
      port "http" {
        to = 80
      }
    }

    service {
      name = "web"
      port = "http"

      tags = [
        "traefik.enable=true",
        "traefik.http.routers.http.rule=Path(`/`)",
      ]

      check {
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
        ports = ["http"]
      }

      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}
