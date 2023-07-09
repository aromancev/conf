job "iam" {
  datacenters = ["dc1"]

  group "iam" {
    network {
      port "http" {
        to = -1
      }
      port "rpc" {
        to = -1
      }
    }

    service {
      name = "iam"
      port = "http"

      tags = [
        "traefik.enable=true",
        "traefik.http.routers.http.rule=PathPrefix(`/api/iam`)",
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
        ports = ["http"]
      }

      env = {
        LISTEN_WEB_ADDRESS = "${NOMAD_PORT_http}"
        LISTEN_RPC_ADDRESS = "${NOMAD_PORT_rpc}"
        WEB_HOST = "confa.io"
        WEB_SCHEME = "http"
        LOG_FORMAT = "json"
        LOG_LEVEL = "info"
        SECRET_KEY = "LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1IY0NBUUVFSUI4Zm1WV2hNZEFvL1VrRE5ONFVHbzhQWXdLeHovbE43bmlsbVlhMktFa2JvQW9HQ0NxR1NNNDkKQXdFSG9VUURRZ0FFVHJNZDBCcjdHT3BFN1VTMWpKN0xiTDBMOHZJaTNOeFJ4blhoT3hEV2FBaGQ0TXhkRjE3ZgpBWTVPR2pKcFBkV0o4VERNUUg3RXM5OFNBQjlwVlJWWmhnPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo="
        PUBLIC_KEY = "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFVHJNZDBCcjdHT3BFN1VTMWpKN0xiTDBMOHZJaQozTnhSeG5YaE94RFdhQWhkNE14ZEYxN2ZBWTVPR2pKcFBkV0o4VERNUUg3RXM5OFNBQjlwVlJWWmhnPT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg=="
        MONGO_HOSTS = "${NOMAD_UPSTREAM_ADDR_mongodb}"
        MONGO_USER = "iam"
        MONGO_PASSWORD = "iam"
        MONGO_DATABASE = "iam"
        BEANSTALK_POOL = "${NOMAD_UPSTREAM_ADDR_beanstalk}"
        BEANSTALK_TUBE_SEND = "sender/send"
        BEANSTALK_TUBE_UPDATE_AVATAR = "confa/update-avatar"
        GOOGLE_API_BASE_URL = ""
        GOOGLE_CLIENT_ID = ""
        GOOGLE_CLIENT_SECRET = ""
      }

      resources {
        cpu    = 100
        memory = 64
      }
    }
  }
}
