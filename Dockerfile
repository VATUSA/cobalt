FROM golang:1.25.6-alpine3.23 AS build
WORKDIR /go/src/github.com/VATUSA/cobalt
COPY src/ ./
RUN go build -o bin/server ./cmd/server.go
RUN go build -o bin/background ./cmd/background.go
RUN go build -o bin/cli ./cmd/cli.go

FROM alpine:3.23 AS app
WORKDIR /app
COPY --from=build /go/src/github.com/VATUSA/cobalt/bin/server ./
COPY --from=build /go/src/github.com/VATUSA/cobalt/bin/background ./
COPY --from=build /go/src/github.com/VATUSA/cobalt/bin/cli ./
COPY sql/ ./sql
ENTRYPOINT ["/app/server"]