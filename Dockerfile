# Use the official Golang image
FROM golang:1.20

# Set the working directory inside the container
WORKDIR /app

# Copy the Go application code
COPY . .

# Build the application
RUN go build -o main .

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./main"]
