job "iam" {
  group "iam" {
    network {
      port "web" {
        to = 80
      }
      port "rpc" {
        to = 8000
      }
    }

    service {
      name = "iam-web"
      port = "web"

      tags = [
        "graph.enable=true",
        "traefik.enable=true",
        "traefik.http.routers.iam.rule=Host(`confa.io`) && PathPrefix(`/api/iam`)",
        "traefik.http.routers.iam.tls=true",
        "traefik.http.routers.iam.tls.certresolver=confa",
        "traefik.http.routers.iam.middlewares=stripprefix-iam",
        "traefik.http.middlewares.stripprefix-iam.stripprefix.prefixes=/api/iam",
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
      name = "iam-rpc"
      port = "rpc"
    }

    task "iam" {
      driver = "docker"

      config {
        image = "confa/iam:latest"
        ports = ["web", "rpc"]
      }

      template {
        data = <<EOH
          LISTEN_WEB_ADDRESS = ":80"
          LISTEN_RPC_ADDRESS = ":8000"
          WEB_HOST = "{{ key "web/host" }}"
          WEB_SCHEME = "{{ key "web/scheme" }}"
          LOG_FORMAT = "json"
          LOG_LEVEL = "info"
          SECRET_KEY = "{{ key "auth/private_key" | base64Encode }}"
          PUBLIC_KEY = "{{ key "auth/public_key" | base64Encode }}"
          MONGO_HOSTS = "{{range $i, $s := service "mongodb" }}{{if ne $i 0}},{{end}}{{$s.Address}}{{end}}"
          MONGO_USER = "iam"
          MONGO_PASSWORD = "iam"
          MONGO_DATABASE = "iam"
          BEANSTALK_POOL = "{{range $i, $s := service "beanstalk" }}{{if ne $i 0}},{{end}}{{$s.Address}}:{{$s.Port}}{{end}}"
          BEANSTALK_TUBE_SEND = "{{ key "beanstalk/tubes/send" }}"
          BEANSTALK_TUBE_UPDATE_AVATAR = "{{ key "beanstalk/tubes/update-avatar" }}"
          GOOGLE_API_BASE_URL = "https://www.googleapis.com"
          GOOGLE_CLIENT_ID = "{{ key "auth/google/client_id" }}"
          GOOGLE_CLIENT_SECRET = "{{ key "auth/google/client_secret" }}"
        EOH

        destination = "secrets/.env"
        env         = true
      }

      resources {
        cpu    = 100
        memory = 32
      }
    }
  }
}
