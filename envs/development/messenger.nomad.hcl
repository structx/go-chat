
job "messenger" {

    datacenters = ["*"]

    type = "service"

    group "trevatk" {
        count = 1

        network {
            mode = "bridge"
            hostname = "messenger.structx.docker"
            port "http" {
                to = 80
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
                sidecar_service {
                    proxy {
                        upstreams {
                            destination_name = "chat-structx-io"
                            local_bind_port = 9090
                        }
                    }
                }
            }

        }

        task "dashboard" {

            driver = "docker"

            config {
                image = "trevatk/messenger:v0.0.1"
                ports = ["http"]
            }

            env {
                REACT_APP_CHAT_SERVICE_ADDR = "http://${NOMAD_UPSTREAM_ADDR_chat_structx_io}"
                REACT_APP_REDUX_STORE_KEY = REDUX_STORE_KEY
            }

            resources {
                cpu    = 500
                memory = 256
            }
        }
    }
}
