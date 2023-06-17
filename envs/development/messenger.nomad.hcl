
job "messenger" {

    datacenters = ["*"]

    type = "service"

    group "trevatk" {
        count = 1

        network {
            mode = "bridge"

            port "http" {
            }
        }

        service {
            name = "messenger-structx-io"
            tags = [
                "reactjs"
            ]
            port = "http"
            provider = "consul"

            connect {
                sidecar_service {}
            }

        }

        task "web-ui" {

            driver = "docker"

            config {
                image = "trevatk/messenger:v0.0.1"
                ports = ["http"]
            }

            env {
                VUE_APP_CHAT_SERVICE_URL = ""
            }

            resources {
                cpu    = 500
                memory = 256
            }
        }
    }
}
