FROM golang:1.26.4-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/exe ./cmd/main.go


FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/exe .
CMD ["./exe"]