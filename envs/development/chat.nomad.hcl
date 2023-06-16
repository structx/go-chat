
job "chat" {

    datacenters = ["*"]

    type = "service"

    group "trevatk" {
        count = 1

        network {
            mode = "bridge"

            port "http" {
            }

            port "grpc" {
            }
        }

        service {
            name = "chat-structx-io"
            tags = [
                "http"
            ]
            port = "http"
            provider = "consul"

            connect {
                sidecar_service {}
            }

            check {
                name = "alive"
                type = "http"
                port = "http"
                path = "/health"
                interval = "1m"
                timeout = "10s"
            }
        }

        volume "migrations" {
            type = "host"
            source = "chat-migrations"
            read_only = true
        }

        volume "db" {
            type = "host"
            source = "chat-db"
            read_only = false
        }

        task "server" {

            driver = "docker"

            config {
                image = "trevatk/go-chat:v0.0.1"
                ports = ["http", "grpc"]
            }

            volume_mount {
                volume = "migrations"
                destination = "/app/migrations"
            }

            volume_mount {
                volume = "db"
                destination = "/app/sqlite"
            }

            env {
                HTTP_SERVER_PORT = "${NOMAD_PORT_http}"
		GRPC_SERVER_PORT = "${NOMAD_PORT_grpc}"
                SQLITE_DSN = "/app/sqlite/chat.db"
                SQLITE_MIGRATIONS_DIR = "/app/migrations"
            }

            resources {
                cpu    = 500
                memory = 256
            }
        }
    }
}
