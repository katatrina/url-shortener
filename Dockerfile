FROM golang:1.26.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /server /server

EXPOSE 8080

ENTRYPOINT ["/server"]