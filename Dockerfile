FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -ldflags="-w -s" -o backups main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/backups .
RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai
CMD ["/app/backups"]