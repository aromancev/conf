job "rtc" {
  group "rtc" {
    network {
      port "web" {
        to = 80
      }
      port "rpc" {
        to = 8000
      }
    }

    service {
      name = "rtc-web"
      port = "web"

      tags = [
        "graph.enable=true",
        "traefik.enable=true",
        "traefik.http.routers.rtc.rule=Host(`confa.io`) && PathPrefix(`/api/rtc`)",
        "traefik.http.routers.rtc.tls=true",
        "traefik.http.routers.rtc.tls.certresolver=confa",
        "traefik.http.routers.rtc.middlewares=stripprefix-rtc",
        "traefik.http.middlewares.stripprefix-rtc.stripprefix.prefixes=/api/rtc",
      ]

      check {
        name     = "alive"
        type     = "http"
        path     = "/health"
        interval = "2s"
        timeout  = "2s"
      }
    }

    service {
      name = "rtc-rpc"
      port = "rpc"
    }

    task "rtc" {
      driver = "docker"

      config {
        image = "confa/rtc:latest"
        ports = ["web", "rpc"]
      }

      template {
        data = <<EOH
          LISTEN_WEB_ADDRESS = ":80"
          LISTEN_RPC_ADDRESS = ":8000"
          LOG_FORMAT = "json"
          LOG_LEVEL = "info"
          PUBLIC_KEY = "{{ key "auth/public_key" | base64Encode }}"
          MONGO_HOSTS = "{{range $i, $s := service "mongodb" }}{{if ne $i 0}},{{end}}{{$s.Address}}{{end}}"
          MONGO_USER = "rtc"
          MONGO_PASSWORD = "rtc"
          MONGO_DATABASE = "rtc"
          BEANSTALK_POOL = "{{range $i, $s := service "beanstalk" }}{{if ne $i 0}},{{end}}{{$s.Address}}:{{$s.Port}}{{end}}"
          BEANSTALK_TUBE_SEND = "{{ key "beanstalk/tubes/send" }}"
          BEANSTALK_TUBE_STORE_EVENT = "{{ key "beanstalk/tubes/store-event" }}"
          BEANSTALK_TUBE_UPDATE_RECORDING_TRACK = "{{ key "beanstalk/tubes/update-recording-track" }}"
          BEANSTALK_TUBE_RECORDING_UPDATE = "{{ key "beanstalk/tubes/recording-update" }}"
          {{range service "tracker-rpc" }}
            TRACKER_RPC_ADDRESS = "{{.Address}}:{{.Port}}"
          {{end}}
          LIVEKIT_KEY = "key"
          LIVEKIT_SECRET = "93d33a06-f209-4239-bd7f-d04d411ae7b2"
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
