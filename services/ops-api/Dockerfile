FROM golang:1.26-alpine AS builder

WORKDIR /src

# Copy go.work and dependencies
COPY go.work ./
COPY pkg/ ./pkg/
COPY services/ops-api/ ./services/ops-api/

# Compile the server from the ops-api directory
WORKDIR /src/services/ops-api
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates && adduser -D appuser
COPY --from=builder /server /server
USER appuser
ENTRYPOINT ["/server"]
