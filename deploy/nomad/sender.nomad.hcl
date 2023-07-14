job "sender" {
  group "sender" {
    service {
      name = "sender"
    }

    task "sender" {
      driver = "docker"

      config {
        image = "confa/sender:latest"
      }

      template {
        data = <<EOH
          LOG_FORMAT = "json"
          LOG_LEVEL = "info"
          IAM_RPC_ADDRESS = "{{range $i, $s := service "iam" }}{{if eq $i 0}},{{end}}{{$s.Address}}:{{$s.Port}}{{end}}"
          MAILERSEND_BASE_URL = "https://api.mailersend.com"
          MAILERSEND_TOKEN = "{{ key sender/mailersend/token }}"
          MAILERSEND_FROM_EMAIL = "noreply@mail.confa.io"
          BEANSTALK_POOL = "{{range $i, $s := service "beanstalk" }}{{if ne $i 0}},{{end}}{{$s.Address}}:{{$s.Port}}{{end}}"
          BEANSTALK_TUBE_SEND = "{{ key "beanstalk/tubes/send" }}"
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
