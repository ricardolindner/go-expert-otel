FROM golang:1.23.6-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /go-weather-api ./cmd/server/main.go


FROM alpine:3.18
WORKDIR /
COPY --from=builder /go-weather-api /go-weather-api
EXPOSE 8080
CMD ["/go-weather-api"]
