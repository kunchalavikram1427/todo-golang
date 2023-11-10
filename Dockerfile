# Use the official Golang image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code and config file to the container
COPY . .
COPY config.json .

# Build the Go application
RUN go build -o todo-app .

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./todo-app"]
