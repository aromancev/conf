job "teleport" {
  group "teleport" {
    constraint {
      attribute = "${meta.ingress_teleport}"
      value     = "true"
    }

    network {
      port "https" {
        static = 443
      }
    }

    volume "teleport" {
      type      = "host"
      read_only = false
      source    = "teleport"
    }

    task "teleport" {
      driver = "docker"

      config {
        image        = "public.ecr.aws/gravitational/teleport-distroless:13.3"
        network_mode = "host"
        # For some reason, teleport can start consuming a lof of CPU.
        cpu_hard_limit = true

        volumes = [
          "local/teleport.yml:/etc/teleport/teleport.yaml",
        ]
      }

      template {
        data = <<EOF
          version: v3
          teleport:
            nodename: {{env "attr.unique.consul.name"}}
            data_dir: /opt/teleport
            log:
              output: stdout
              severity: INFO
              format:
                output: json
          auth_service:
            enabled: true
            proxy_listener_mode: multiplex
            authentication:
              type: github
          proxy_service:
            enabled: true
            web_listen_addr: 0.0.0.0:443
            public_addr: teleport.{{ key "web/host" }}:443
            https_keypairs: []
            https_keypairs_reload_interval: 0s
            acme:
              enabled: true
              email: {{ key "tls/confa/email" }}
          app_service:
              enabled: true
              apps:
                - name: consul
                  uri: http://localhost:8500
                  public_addr: consul.teleport.{{ key "web/host" }}
                - name: nomad
                  uri: http://localhost:4646
                  public_addr: nomad.teleport.{{ key "web/host" }}
                - name: traefik
                  uri: http://traefik.service.consul:8000/dashboard/
                  public_addr: traefik.teleport.{{ key "web/host" }}
                {{range service "minio-console" }}
                - name: minio
                  uri: http://{{.Address}}:{{.Port}}
                  public_addr: minio.teleport.{{ key "web/host" }}
                {{end}}
          ssh_service:
            enabled: false
        EOF
        destination = "local/teleport.yml"
      }

      volume_mount {
        volume      = "teleport"
        destination = "/opt/teleport"
        read_only   = false
      }

      resources {
        cpu    = 500
        memory = 256
        memory_max = 512
      }
    }
  }
}
