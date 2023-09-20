FROM golang:1.21 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o backend

FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/backend /backend
CMD ["/backend", "api"]