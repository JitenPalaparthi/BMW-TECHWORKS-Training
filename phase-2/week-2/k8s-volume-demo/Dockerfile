FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app

FROM alpine:3.22
#RUN addgroup -S app && adduser -S app -G app && mkdir -p /data && chown -R app:app /data
RUN mkdir -p /data
COPY --from=builder /app /app
# USER app
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/app"]
