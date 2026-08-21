FROM golang:1.25.10-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o spark-apply .

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache docker-cli
COPY --from=builder /app/spark-apply .
EXPOSE 8080
VOLUME /var/run/docker.sock
CMD ["./spark-apply"]
