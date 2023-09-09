job "tracker" {
  group "tracker" {
    network {
      port "rpc" {
        to = 8000
      }
    }

    service {
      name = "tracker-rpc"
      port = "rpc"
    }

    volume "tracker" {
      type      = "host"
      read_only = false
      source    = "tracker"
    }

    task "tracker" {
      driver = "docker"
      user = "1003:1003" # Pre-defined well-known global constant. See terraform configuration.

      config {
        image = "confa/tracker:latest"
        ports = ["rpc"]
      }

      template {
        data = <<EOH
          LISTEN_RPC_ADDRESS = ":8000"
          LOG_FORMAT = "json"
          LOG_LEVEL = "info"
          PUBLIC_KEY = "{{ key "auth/public_key" | base64Encode }}"
          TMP_DIR = "/opt/tracker"
          BEANSTALK_POOL = "{{range $i, $s := service "beanstalk" }}{{if ne $i 0}},{{end}}{{$s.Address}}:{{$s.Port}}{{end}}"
          BEANSTALK_TUBE_PROCESS_TRACK = "{{ key "beanstalk/tubes/process-track" }}"
          BEANSTALK_TUBE_STORE_EVENT = "{{ key "beanstalk/tubes/store-event" }}"
          BEANSTALK_TUBE_UPDATE_RECORDING_TRACK = "{{ key "beanstalk/tubes/update-recording-track" }}"
          {{range service "livekit-ws" }}
            LIVEKIT_URL = "ws://{{.Address}}:{{.Port}}"
          {{end}}
          LIVEKIT_KEY = "key"
          LIVEKIT_SECRET = "93d33a06-f209-4239-bd7f-d04d411ae7b2"
          {{range service "minio" }}
            STORAGE_HOST = "{{.Address}}:{{.Port}}"
          {{end}}
          STORAGE_SCHEME = "http"
          STORAGE_REGION = "stub"
          STORAGE_ACCESS_KEY = "{{ key "storage/access-key" }}"
          STORAGE_SECRET_KEY = "{{ key "storage/secret-key" }}"
          STORAGE_BUCKET_TRACK_RECORDS = "{{ key "storage/buckets/confa-tracks-internal" }}"
        EOH
        destination = "secrets/.env"
        env         = true
      }

      volume_mount {
        volume      = "tracker"
        destination = "/opt/tracker"
        read_only   = false
      }

      resources {
        cpu    = 300
        memory = 256
        memory_max = 1024
      }
    }
  }
}
