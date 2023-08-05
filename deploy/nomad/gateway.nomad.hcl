job "gateway" {
  group "gateway" {
    network {
      port "web" {
        to = 80
      }
    }

    service {
      name = "gateway"
      port = "web"

      tags = [
        "traefik.enable=true",
        "traefik.http.routers.gateway.rule=Host(`confa.io`) && PathPrefix(`/api/query`)",
        "traefik.http.routers.gateway.tls=true",
        "traefik.http.routers.gateway.tls.certresolver=confa",
        "traefik.http.routers.gateway.middlewares=stripprefix-gateway",
        "traefik.http.middlewares.stripprefix-gateway.stripprefix.prefixes=/api",
      ]

      check {
        name     = "alive"
        type     = "http"
        path     = "/query"
        interval = "2s"
        timeout  = "2s"
      }
    }

    task "gateway" {
      driver = "docker"

      config {
        image = "confa/gateway:latest"
        ports = ["web"]
      }

      template {
        data = <<EOH
          LISTEN_WEB_ADDRESS = ":80"
          LOG_FORMAT = "json"
          LOG_LEVEL = "info"
          SCHEMA_UPDATE_INTERVAL_S = "10"
          SERVICES = "{{ range services }}{{ if .Tags | contains "graph.enable=true" }}{{ range service .Name }}http://{{ .Address }}:{{ .Port }}/graph,{{ end }}{{ end }}{{ end }}"
        EOH

        destination = "secrets/.env"
        env         = true
      }

      resources {
        cpu    = 100
        memory = 64
      }
    }
  }
}
