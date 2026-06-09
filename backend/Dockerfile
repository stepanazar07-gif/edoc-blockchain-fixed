FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/edoc-server ./cmd/server

FROM alpine:3.22

WORKDIR /app
RUN mkdir -p /app/uploads

COPY --from=builder /out/edoc-server /app/edoc-server

ENV API_PORT=8080
ENV P2P_ADDR=0.0.0.0:8001
ENV UPLOAD_DIR=/app/uploads

EXPOSE 8080 8001

ENTRYPOINT ["/app/edoc-server"]
