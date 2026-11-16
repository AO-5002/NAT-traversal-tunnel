FROM golang:alpine3.24 AS builder
WORKDIR /build
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o relay ./cmd/relay

FROM gcr.io/distroless/static
WORKDIR /app
COPY --from=builder /build/relay ./relay
EXPOSE 8080 8080
CMD ["/app/relay"]