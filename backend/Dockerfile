FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/main main.go

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /out/main /app/main
COPY resource /app/resource
COPY manifest/config /app/manifest/config

EXPOSE 8080

CMD ["./main"]
