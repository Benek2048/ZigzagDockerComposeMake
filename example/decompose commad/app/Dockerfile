FROM golang:1.21-alpine

WORKDIR /app

# Copy everything
COPY . .

# Download dependencies and build
RUN go get github.com/redis/go-redis/v9 && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o main .

EXPOSE 7001

CMD ["./main"]