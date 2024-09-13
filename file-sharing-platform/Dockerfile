# Use the official Golang image as the base
FROM golang:1.23

# Set the working directory in the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o main .

# Expose the port on which the Go app will run
EXPOSE 8080

# Command to run the Go app
CMD ["./main"]
