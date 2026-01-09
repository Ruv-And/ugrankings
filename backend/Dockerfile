FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY backend/go.mod .
RUN go mod download
COPY backend/. .
RUN go build -o /bin/api

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /bin/api /bin/api
EXPOSE 8080
CMD ["/bin/api"]
