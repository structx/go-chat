#!/bin/bash

# build docker images
docker build -t structx/chat:v0.1.0 -f ../docker/server.Dockerfile .
docker build -t structx/chat-migrate:v0.1.0 -f ../docker/migrate.Dockerfile .

# run nomad jobs
nomad job run ../env/production/chat.nomad.hcl
nomad job run ../env/production/messenger.nomad.hcl