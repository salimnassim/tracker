FROM golang:latest as builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -v -o ./tracker ./cmd

FROM scratch
COPY --from=builder /app/tracker /app/tracker
COPY --from=builder /app/templates /app/templates
CMD ["/app/tracker"]