FROM golang:1.23-alpine AS builder
WORKDIR /app
# We'll initialize the module during the build if you haven't yet
COPY . .
RUN go mod init chidaucf && go mod tidy
RUN go build -o server .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
