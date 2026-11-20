FROM golang:1.23-alpine AS builder

WORKDIR /src
RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .
RUN go get github.com/gorilla/websocket@v1.5.3
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/flexgram-server ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/flexgram-server /app/flexgram-server
COPY internal/database/migrations /app/migrations
EXPOSE 8080
ENTRYPOINT ["/app/flexgram-server"]
