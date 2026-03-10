FROM --platform=$BUILDPLATFORM golang:1.24.3-bookworm AS builder
WORKDIR /app
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o server .

FROM alpine:3.21
WORKDIR /app
RUN apk add --no-cache ca-certificates && \
    addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=builder /app/server /app/server

USER appuser

EXPOSE 8080

ENTRYPOINT ["./server"]
