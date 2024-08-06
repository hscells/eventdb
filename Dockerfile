FROM golang:1.22

WORKDIR /usr/src/eventdb

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY *.go ./
COPY cmd/eventdb/eventdb-server.go ./cmd/eventdb/eventdb-server.go
RUN go build -o /usr/bin/eventdb-server ./cmd/eventdb/eventdb-server.go

ENV GIN_MODE=release

CMD ["eventdb-server", "settings.toml"]