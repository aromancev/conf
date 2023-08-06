locals {
  port_ws = 7880
  port_rtc = 7881
  port_turn_udp = 3478
  port_turn_tls = 5349
}

job "livekit" {
  group "livekit" {
    constraint {
      attribute = "${meta.ingress_sfu}"
      value     = "true"
    }

    network {
      port "ws" {
        static = local.port_ws
      }
      port "turn" {
        static = local.port_turn_tls
      }
    }

    service {
      name = "livekit-ws"
      port = "ws"

      tags = [
        "traefik.enable=true",
        "traefik.http.routers.livekit.rule=Host(`sfu.confa.io`)",
        "traefik.http.routers.livekit.tls=true",
        "traefik.http.routers.livekit.tls.certresolver=confa",
      ]
    }

    service {
      name = "livekit-turn"
      port = "turn"

      tags = [
        "traefik.enable=true",
        "traefik.tcp.routers.livekit.rule=HostSNI(`turn.confa.io`)",
        "traefik.tcp.routers.livekit.tls=true",
        "traefik.tcp.routers.livekit.tls.certresolver=confa",
      ]
    }

    task "livekit" {
      driver = "docker"

      config {
        image        = "livekit/livekit-server:v1.4"
        network_mode = "host"
        args = [
          "--config",
          "/etc/livekit/livekit.yml",
        ]

        volumes = [
          "local/livekit.yml:/etc/livekit/livekit.yml",
        ]
      }

      template {
        data = <<EOF
          # main TCP port for RoomService and RTC endpoint
          # for production setups, this port should be placed behind a load balancer with TLS
          port: ${local.port_ws}

          # WebRTC configuration
          rtc:
            # UDP ports to use for client traffic.
            # this port range should be open for inbound traffic on the firewall
            port_range_start: 50000
            port_range_end: 60000
            # when set, LiveKit enable WebRTC ICE over TCP when UDP isn't available
            # this port *cannot* be behind load balancer or TLS, and must be exposed on the node
            # WebRTC transports are encrypted and do not require additional encryption
            # only 80/443 on public IP are allowed if less than 1024
            tcp_port: ${local.port_rtc}

          # Signal Relay
          # since v1.4.0, a more reliable, psrpc based signal relay is available
          # this gives us the ability to reliably proxy messages between a signal server and RTC node
          signal_relay:
            # disabled by default. will be enabled by default in future versions
            enabled: true

          # API key / secret pairs.
          # Keys are used for JWT authentication, server APIs would require a keypair in order to generate access tokens
          # and make calls to the server
          keys:
            key: 93d33a06-f209-4239-bd7f-d04d411ae7b2

          # Logging config
          logging:
            # log level, valid values: debug, info, warn, error
            level: info
            # log level for pion, default error
            pion_level: info
            # when set to true, emit json fields
            json: true
            # for production setups, enables sampling algorithm
            # https://github.com/uber-go/zap/blob/master/FAQ.md#why-sample-application-logs
            sample: true

          # turn server
          turn:
            # Uses TLS. Requires cert and key pem files by either:
            # - using turn.secretName if deploying with our helm chart, or
            # - setting LIVEKIT_TURN_CERT and LIVEKIT_TURN_KEY env vars with file locations, or
            # - using cert_file and key_file below
            # defaults to false
            enabled: true
            # defaults to 3478 - recommended to 443 if not running HTTP3/QUIC server
            # only 53/80/443 are allowed if less than 1024
            udp_port: ${local.port_turn_udp}
            # defaults to 5349 - if not using a load balancer, this must be set to 443
            tls_port: ${local.port_turn_tls}
            external_tls: true
            # needs to match tls cert domain
            domain: turn.confa.io
        EOF
        destination = "local/livekit.yml"
      }

      resources {
        cpu    = 1000
        memory = 512
      }
    }
  }
}
