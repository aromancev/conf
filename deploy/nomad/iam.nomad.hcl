job "iam" {
  datacenters = ["dc1"]

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
      name = "iam"
      port = "web"

      tags = [
        "traefik.enable=true",
        "traefik.http.routers.iam.rule=PathPrefix(`/api/iam`)",
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
          WEB_HOST = "confa.io"
          WEB_SCHEME = "http"
          LOG_FORMAT = "json"
          LOG_LEVEL = "info"
          SECRET_KEY = "LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1IY0NBUUVFSUI4Zm1WV2hNZEFvL1VrRE5ONFVHbzhQWXdLeHovbE43bmlsbVlhMktFa2JvQW9HQ0NxR1NNNDkKQXdFSG9VUURRZ0FFVHJNZDBCcjdHT3BFN1VTMWpKN0xiTDBMOHZJaTNOeFJ4blhoT3hEV2FBaGQ0TXhkRjE3ZgpBWTVPR2pKcFBkV0o4VERNUUg3RXM5OFNBQjlwVlJWWmhnPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo="
          PUBLIC_KEY = "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFVHJNZDBCcjdHT3BFN1VTMWpKN0xiTDBMOHZJaQozTnhSeG5YaE94RFdhQWhkNE14ZEYxN2ZBWTVPR2pKcFBkV0o4VERNUUg3RXM5OFNBQjlwVlJWWmhnPT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg=="
          MONGO_HOSTS = "{{range $i, $s := service "mongodb" }}{{if ne $i 0}},{{end}}{{$s.Address}}{{end}}"
          MONGO_USER = "iam"
          MONGO_PASSWORD = "iam"
          MONGO_DATABASE = "iam"
          BEANSTALK_POOL = "{{range $i, $s := service "beanstalk" }}{{if ne $i 0}},{{end}}{{$s.Address}}:{{$s.Port}}{{end}}"
          BEANSTALK_TUBE_SEND = "sender/send"
          BEANSTALK_TUBE_UPDATE_AVATAR = "confa/update-avatar"
          GOOGLE_API_BASE_URL = "https://www.googleapis.com"
          GOOGLE_CLIENT_ID = "stub"
          GOOGLE_CLIENT_SECRET = "stub"
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
