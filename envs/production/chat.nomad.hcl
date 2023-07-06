
job "chat" {

    datacenters = ["us-mountain-1"]

    type = "service"

    group "trevatk" {
        count = 1

        network {
            mode = "bridge"
            hostname = "chat.structx.io"
            port "http" {
            }

            port "grpc" {
            }
        }

        service {
            name = "chat-api"
            tags = [
                "traefik.enable=true",
                "traefik.http.routers.chat-api.entryPoints=websecure",
                "traefik.http.routers.chat-api.rule=Host(`chat.structx.io`)",
                "traefik.http.routers.chat-api.tls=true",
                "treafik.http.routers.tls.certresolver=myresolver",
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

        service {
            name = "chat-grpc"
            tags = [
                "traefik.enable=true",
                "traefik.http.routers.chat-grpc.entryPoints=grpc",
                "traefik.http.routers.chat-grpc.rule=Host(`chat.structx.io`)",
                "traefik.http.routers.chat-grpc.tls=true",
                "treafik.http.routers.tls.certresolver=myresolver",
                "grpc"
            ]
            port = "grpc"
            provider = "consul"

            connect {
                sidecar_service {}
            }
        }

        volume "db" {
            type = "host"
            source = "chat-db"
            read_only = false
        }

        volume "certs" {
            type = "host"
            source = "chat-certs"
            read_only = true
        }

        task "migrate" {

            driver = "docker"
            
            config {
                image = "structx/chat-migrate:v0.1.0"
            }

            lifecycle {
                hook = "prestart"
                sidecar = false
            }
        }

        task "server" {

            driver = "docker"

            config {
                image = "trevatk/go-chat:v0.0.1"
                ports = ["http", "grpc"]
            }

            volume_mount {
                volume = "db"
                destination = "/app/sqlite"
            }

            volume_mount {
                volume = "certs"
                destination = "/app/certs"
            }

            env {
                HTTP_SERVER_PORT = "${NOMAD_PORT_http}"
		        GRPC_SERVER_PORT = "${NOMAD_PORT_grpc}"
                SQLITE_DSN = "/app/sqlite/chat.db"
                SQLITE_MIGRATIONS_DIR = "/app/migrations"
                LOG_LEVEL = "production"
                ALLOWED_ORIGINS = "https://messenger.structx.io"
                JWT_PRIVATE_KEY = "/app/certs/keyPair.pem"
            }

            resources {
                cpu    = 500
                memory = 256
            }
        }
    }
}
