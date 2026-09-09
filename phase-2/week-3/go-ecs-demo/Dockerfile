FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./

RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build -o app .

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]