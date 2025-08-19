FROM golang:1.21-alpine AS builder
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY pkg/ pkg/
RUN CGO_ENABLED=0 GOOS=linux go build -o controller ./cmd/operator

FROM alpine:3.18
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=builder /workspace/controller .
USER 1001
ENTRYPOINT ["/controller"]
