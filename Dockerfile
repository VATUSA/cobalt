FROM golang:1.25.6-alpine3.23 AS build
WORKDIR /go/src/github.com/VATUSA/cobalt
COPY src/ ./
RUN go build -o bin/cobalt-server ./cmd/server.go
RUN go build -o bin/cobalt-background ./cmd/background.go
RUN go build -o bin/cobalt-cli ./cmd/cli.go

FROM alpine:3.23 AS app
WORKDIR /app
COPY --from=build /go/src/github.com/VATUSA/cobalt/bin/cobalt-server ./
COPY --from=build /go/src/github.com/VATUSA/cobalt/bin/cobalt-background ./
COPY --from=build /go/src/github.com/VATUSA/cobalt/bin/cobalt-cli ./
COPY sql/ ./sql
ENTRYPOINT ["/app/cobalt-server"]