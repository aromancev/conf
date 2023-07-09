job "traefik" {
  type = "system"

  group "traefik" {
    network {
      port "gateway" {
        static = 80
      }
      port "api" {
        static = 8000
      }
    }

    service {
      name = "traefik"

      check {
        name     = "alive"
        type     = "tcp"
        port     = "gateway"
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
          logLevel: "INFO"

          log:
            format: "json"

          accessLog:
            format: "json"

          entryPoints:
            gateway:
              address: ":80"
            traefik:
              address: ":8000"

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
