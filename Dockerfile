FROM golang:1.26.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

ARG GEOIP_MONTH=2026-08
RUN mkdir -p /geoip && \
    curl -fsSL "https://download.db-ip.com/free/dbip-country-lite-${GEOIP_MONTH}.mmdb.gz" \
    | gunzip > /geoip/dbip-country-lite.mmdb

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /server /server
COPY --from=builder /geoip/dbip-country-lite.mmdb /geoip/dbip-country-lite.mmdb

EXPOSE 8080

ENTRYPOINT ["/server"]