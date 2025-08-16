FROM golang:1.23.6-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /go-weather-input ./cmd/server/main.go


FROM alpine:3.18
WORKDIR /
COPY --from=builder /go-weather-input /go-weather-input
EXPOSE 8080
CMD ["/go-weather-input"]
