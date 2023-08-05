job "traefik" {
  group "traefik" {
    constraint {
      attribute = "${meta.ingress_web}"
      value     = "true"
    }

    network {
      port "https" {
        static = 443
      }
      port "http" {
        static = 80
      }
      port "traefik" {
        static = 8000
      }
    }

    service {
      name = "traefik"

      check {
        name     = "alive"
        type     = "tcp"
        port     = "traefik"
        interval = "10s"
        timeout  = "2s"
      }
    }

    volume "traefik" {
      type      = "host"
      read_only = false
      source    = "traefik"
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
            http:
              address: ":80"
              http:
                redirections:
                  entryPoint:
                    to: https
                    scheme: https
            https:
              address: ":443"
            traefik:
              address: ":8000"

          certificatesResolvers:
            confa:
              acme:
                email: {{ key "tls/confa/email" }}
                storage: /etc/traefik/acme/confa.json
                httpChallenge:
                  entryPoint: http

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

      volume_mount {
        volume      = "traefik"
        destination = "/etc/traefik"
        read_only   = false
      }

      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}
