job "mongodb" {
  type = "system"
  datacenters = ["dc1"]

  group "mongodb" {
    network {
      port "db" {
        to = 27017
      }
    }

    volume "mongodb" {
      type      = "host"
      read_only = false
      source    = "mongodb"
    }

    restart {
      attempts = 10
      interval = "5m"
      delay    = "25s"
      mode     = "delay"
    }

    service {
      name = "mongodb"
      port = "db"

      check {
        name     = "alive"
        type     = "tcp"
        interval = "10s"
        timeout  = "2s"
      }
      check {
        name     = "ready"
        type     = "script"
        interval = "10s"
        timeout  = "10s"
        task     = "mongodb"
        command  = "mongo"
        args     = [
          "--eval",
          "'db.runCommand(\"ping\").ok'",
          "localhost:27017",
          "--quiet",
        ]
      }
    }

    task "mongodb" {
      driver = "docker"
      user = "1001:1001"

      config {
        image = "mongodb/mongodb-community-server:4.4-ubuntu2004"
        command = "mongod"
        args = [
          "--dbpath",
          "/opt/mongodb/data",
          "--replSet",
          "rs",
          "--keyFile",
          "/opt/mongodb/repl.key",
        ]
        ports = ["db"]
      }

      env = {
        MONGODB_INITDB_ROOT_USERNAME = "mongodb"
        MONGODB_INITDB_ROOT_PASSWORD = "mongodb"
      }

      volume_mount {
        volume      = "mongodb"
        destination = "/opt/mongodb"
        read_only   = false
      }

      resources {
        cpu    = 500
        memory = 512
      }
    }
  }
}
