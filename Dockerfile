FROM golang:1.16 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server cmd/server/main.go

FROM scratch
COPY --from=builder /app/server /bin/app
CMD ["/bin/app"]