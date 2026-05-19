FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o mxcmdb cmd/server/main.go

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/mxcmdb .
COPY --from=builder /app/configs/config.example.yaml ./configs/config.yaml

EXPOSE 10030
CMD ["./mxcmdb"]
