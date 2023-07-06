FROM golang:1.20-bullseye as builder

WORKDIR /app/src

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 go build -o /app/bin/migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian11

COPY --from=builder /app/bin/ /app/bin/

COPY migrations /app/migrations

VOLUME [ "/app/sqlite" ]

ENTRYPOINT ["/app/bin/migrate"]
