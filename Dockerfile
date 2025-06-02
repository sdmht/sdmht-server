# 构建阶段
FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o sdmht-server ./server.go

# 运行阶段
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/sdmht-server .
EXPOSE 8000
ENTRYPOINT ["./sdmht-server"]
