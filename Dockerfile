FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy && go mod download

COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -o cloudflaretinyurl .

# Final stage
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY test /app/test

COPY --from=builder /app/cloudflaretinyurl .

EXPOSE 8080

CMD ["./cloudflaretinyurl"]
