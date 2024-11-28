# Start from the official Golang image
FROM golang:1.20-alpine

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main .

# Expose port 8080
EXPOSE 8080

# Set environment variables (adjust as needed)
ENV DB_USER=go_user
ENV DB_PASS=qwedsa@123
ENV DB_HOST=db
ENV DB_NAME=toronto_time_db

# Run the executable
CMD ["./main"]