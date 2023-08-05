job "confa" {
  group "confa" {
    network {
      port "web" {
        to = 80
      }
      port "rpc" {
        to = 8000
      }
    }

    service {
      name = "confa"
      port = "web"

      tags = [
        "graph.enable=true",
      ]
    }

    task "confa" {
      driver = "docker"

      config {
        image = "confa/confa:latest"
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
          PUBLIC_KEY = "{{ key "auth/public_key" | base64Encode }}"
          MONGO_HOSTS = "{{range $i, $s := service "mongodb" }}{{if ne $i 0}},{{end}}{{$s.Address}}{{end}}"
          MONGO_USER = "confa"
          MONGO_PASSWORD = "confa"
          MONGO_DATABASE = "confa"
          BEANSTALK_POOL = "{{range $i, $s := service "beanstalk" }}{{if ne $i 0}},{{end}}{{$s.Address}}:{{$s.Port}}{{end}}"
          BEANSTALK_TUBE_SEND = "{{ key "beanstalk/tubes/send" }}"
          BEANSTALK_TUBE_UPDATE_AVATAR = "{{ key "beanstalk/tubes/update-avatar" }}"
          BEANSTALK_TUBE_START_RECORDING = "{{ key "beanstalk/tubes/start-recording" }}"
          BEANSTALK_TUBE_STOP_RECORDING = "{{ key "beanstalk/tubes/stop-recording" }}"
          BEANSTALK_TUBE_RECORDING_UPDATE = "{{ key "beanstalk/tubes/recording-update" }}"
          {{range service "minio" }}
            STORAGE_HOST = "{{.Address}}:{{.Port}}"
          {{end}}
          STORAGE_ACCESS_KEY = "minio"
          STORAGE_SECRET_KEY = "miniominio"
          STORAGE_PUBLIC_URL = "/api/storage"
          STORAGE_BUCKET_USER_UPLOADS = "{{ key "storage/buckets/user-uploads" }}"
          STORAGE_BUCKET_USER_PUBLIC = "{{ key "storage/buckets/user-public" }}"
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
