FROM golang:1.26.6 AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -o ./tracker ./cmd

FROM scratch
COPY --from=builder /app/tracker /app/tracker
CMD ["/app/tracker"]