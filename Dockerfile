FROM golang:1.20-alpine AS build

WORKDIR /build
COPY . .
RUN go mod download
EXPOSE 8031
RUN go build -o /build/alertmanager-mqtt-bridge ./cmd/main.go

FROM alpine:latest

WORKDIR /

COPY --from=build /build/alertmanager-mqtt-bridge /alertmanager-mqtt-bridge
EXPOSE 8031
USER nonroot:nonroot

ENTRYPOINT ["/alertmanager-mqtt-bridge"]