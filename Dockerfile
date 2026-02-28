# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Build backend
FROM golang:1.21 AS backend-builder
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/studygolang github.com/studygolang/studygolang/cmd/studygolang

# Stage 3: Production image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /data/www/studygolang

COPY --from=backend-builder /app/bin/studygolang ./bin/studygolang
COPY --from=backend-builder /app/data ./data
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
COPY template ./template
COPY static ./static
COPY config/env.sample.ini ./config/env.ini

RUN mkdir -p log pid

EXPOSE 8088
ENTRYPOINT ["bin/studygolang"]
