FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/warpdns .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/warpdns /warpdns
COPY config.example.toml /etc/warpdns/config.toml
EXPOSE 1053/udp
USER nonroot:nonroot
ENTRYPOINT ["/warpdns"]
CMD ["-config", "/etc/warpdns/config.toml"]
