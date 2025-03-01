# Use Golang base image
FROM golang:1.18

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project into the container
COPY . .

# Build the Go application
RUN go build -o cloudflaretinyurl .

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./cloudflaretinyurl"]
