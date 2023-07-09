job "mongodb" {
  group "mongodb" {
    network {
      port "db" {
        # Port has to be static to initiative replica set. See https://www.mongodb.com/docs/manual/reference/command/replSetInitiate/#mongodb-dbcommand-dbcmd.replSetInitiate.
        static = 27017
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
        timeout  = "5s"
        task     = "mongodb"
        command  = "mongo"
        args     = [
          "--quiet",
          "--eval",
          "'db.runCommand(\"ping\").ok'",
        ]
      }
    }

    task "mongodb" {
      driver = "docker"
      user = "1001:1001"

      config {
        image = "mongodb/mongodb-community-server:4.4-ubuntu2004"
        # Host network is required to access Consul DNS on the host machine.
        # Alternatively, we could dance with docker DNS forwarding. See: https://github.com/hashicorp/nomad/issues/12894.
        network_mode = "host"
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
        MONGODB_INITDB_ROOT_USERNAME = "admin"
        MONGODB_INITDB_ROOT_PASSWORD = "admin"
      }

      volume_mount {
        volume      = "mongodb"
        destination = "/opt/mongodb"
        read_only   = false
      }

      resources {
        cpu    = 500
        memory = 256
      }
    }
  }
}
