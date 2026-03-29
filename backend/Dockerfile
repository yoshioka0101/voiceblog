# ---- build ----
FROM golang:1.26.1-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api

# ---- runtime ----
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /api /api
EXPOSE 8080
CMD ["/api"]
