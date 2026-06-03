# Build from the server directory:
# docker build -f app-api/Dockerfile -t shennong-app-api .

FROM golang:1.21-alpine AS builder

WORKDIR /src

COPY app-api/go.mod app-api/go.sum ./app-api/
COPY common/go.mod common/go.sum ./common/
COPY service/go.mod service/go.sum ./service/

WORKDIR /src/app-api
RUN go mod download

WORKDIR /src
COPY app-api ./app-api
COPY common ./common
COPY service ./service

WORKDIR /src/app-api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/app-api cmd.go

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/app-api ./app-api
COPY app-api/configs ./configs

RUN mkdir -p ./logs

ENV TZ=Asia/Shanghai

EXPOSE 8391

CMD ["./app-api"]
