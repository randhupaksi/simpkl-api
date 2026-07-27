FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/simpkl-api ./cmd/api

FROM alpine:3.22
RUN addgroup -S simpkl && adduser -S simpkl -G simpkl
WORKDIR /app
COPY --from=builder /out/simpkl-api /app/simpkl-api
COPY docs /app/docs
RUN mkdir -p /app/storage/private && chown -R simpkl:simpkl /app
USER simpkl
EXPOSE 8080
ENTRYPOINT ["/app/simpkl-api"]
