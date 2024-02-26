FROM golang:1.21 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build .

FROM gcr.io/distroless/static

WORKDIR /app

COPY --from=builder /build/instatus-cluster-monitor .

ENTRYPOINT ["/app/instatus-cluster-monitor"]
