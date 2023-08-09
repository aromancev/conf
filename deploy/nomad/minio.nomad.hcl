job "minio" {
  group "minio" {
    network {
      port "web" {
        to = 80
      }
    }

    volume "minio" {
      type      = "host"
      read_only = false
      source    = "minio"
    }

    restart {
      attempts = 10
      interval = "5m"
      delay    = "25s"
      mode     = "delay"
    }

    service {
      name = "minio"
      port = "web"

      tags = [
        "traefik.enable=true",
        "traefik.http.routers.minio.rule=Host(`storage.confa.io`)",
        "traefik.http.routers.minio.tls=true",
        "traefik.http.routers.minio.tls.certresolver=confa",
      ]

      check {
        name     = "alive"
        type     = "http"
        path     = "/minio/health/live"
        interval = "10s"
        timeout  = "2s"
      }
    }

    task "minio" {
      driver = "docker"
      user = "1002:1002" # Pre-defined well-known global constant. See terraform configuration.

      config {
        image = "minio/minio:RELEASE.2022-04-16T04-26-02Z"
        command = "server"
        args = [
          "/opt/minio/data",
          "--address",
          ":80",
        ]
        ports = ["web"]
      }

      template {
        data = <<EOH
          MINIO_ROOT_USER = "{{ key "storage/access-key" }}"
          MINIO_ROOT_PASSWORD = "{{ key "storage/secret-key" }}"
        EOH 
        destination = "secrets/.env"
        env         = true
      }

      volume_mount {
        volume      = "minio"
        destination = "/opt/minio"
        read_only   = false
      }

      resources {
        cpu    = 500
        memory = 128
      }
    }
  }
}
