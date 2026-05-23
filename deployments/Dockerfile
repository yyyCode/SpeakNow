FROM golang:1.23-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /speaknow ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /speaknow .
COPY configs/config.yaml ./configs/config.yaml
COPY web ./web
EXPOSE 8080
CMD ["./speaknow", "-config", "configs/config.yaml"]
