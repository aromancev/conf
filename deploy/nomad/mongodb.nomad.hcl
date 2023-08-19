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
      user = "1001:1001" # Pre-defined well-known global constant. See terraform configuration.

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
        volumes = [
          "local/hosts:/etc/hosts",
        ]
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

      # Always resolve mongodb Consul address to the current host to avoid deadlock.
      # Consul does not return unhealthy services, so it will not resolve before the process starts. 
      # The process cannot start if the node is unreachable.
      template {
        data = <<EOF
          {{env "attr.unique.network.ip-address"}} mongodb.service.consul
        EOF
        destination = "local/hosts"
      }

      resources {
        cpu    = 256
        memory = 256
        memory_max = 512
      }
    }
  }
}
