# Use the official Golang image as a base
FROM golang:1.22

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules file and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8081

# Command to run the executable
CMD ["./main", "-up--all"]
