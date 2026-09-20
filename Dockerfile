# stage 1 - build
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ysmtp .

# stage 2 - runtime with alpine
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/ysmtp .
EXPOSE 2525
ENTRYPOINT ["./ysmtp"]
CMD ["start", "-d", "0.0.0.0", "-p", "2525"]
